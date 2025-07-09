#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/wait.h>
#include <sys/time.h>
#include <errno.h>
#include <fcntl.h>
#include <signal.h>
#include <pwd.h>
#include <grp.h>
#include <time.h>

#ifdef JSONC
 #include <json-c/json.h>
#else
 #include <json/json.h>
#endif

#include "command.h"
#include "ispappcwmp.h"
#include "log.h"
#include "time.h"

// Safe command whitelist - only allow specific commands for security
const char *safe_command_patterns[] = {
	"ping",
	"ping6", 
	"traceroute",
	"traceroute6",
	"nslookup",
	"dig",
	"curl",
	"wget",
	"iperf",
	"iperf3",
	"speedtest",
	"uci",
	"cat /proc/",
	"cat /sys/",
	"ls",
	"ps",
	"top",
	"free",
	"df",
	"uptime",
	"date",
	"whoami",
	"id",
	"uname",
	"ifconfig",
	"ip",
	"route",
	"netstat",
	"ss",
	"iwconfig",
	"iwlist",
	"logread",
    "dmesg",
    "log",
    "logcat",
    "ethtool",
    "spectraltool",
    "iw",
    "iwinfo",
    "luci-reload",
    "/etc/init.d/"
	NULL
};

static int command_initialized = 0;

int command_init(void)
{
	if (command_initialized) {
		return 0;
	}
	
	log_message(NAME, L_NOTICE, "command execution module initialized\n");
	command_initialized = 1;
	return 0;
}

void command_cleanup(void)
{
	if (!command_initialized) {
		return;
	}
	
	log_message(NAME, L_NOTICE, "command execution module cleanup\n");
	command_initialized = 0;
}

// Validate if command is in the safe list
int command_validate_safe(const char *command)
{
	if (!command || strlen(command) == 0) {
		return 0;
	}
	
	// Check if command length is reasonable
	if (strlen(command) > COMMAND_MAX_LENGTH) {
		log_message(NAME, L_WARNING, "command too long: %zu characters\n", strlen(command));
		return 0;
	}
	
	// Check against whitelist
	for (int i = 0; safe_command_patterns[i] != NULL; i++) {
		if (strncmp(command, safe_command_patterns[i], strlen(safe_command_patterns[i])) == 0) {
			return 1;
		}
	}
	
	log_message(NAME, L_WARNING, "command not in whitelist: %s\n", command);
	return 0;
}

// Sanitize path to prevent directory traversal
int command_sanitize_path(char *path)
{
	if (!path) {
		return 0;
	}
	
	// Remove dangerous sequences
	char *dangerous[] = {"../", "..\\", "/..", "\\..", NULL};
	for (int i = 0; dangerous[i] != NULL; i++) {
		char *pos = strstr(path, dangerous[i]);
		while (pos != NULL) {
			// Replace with underscore
			memset(pos, '_', strlen(dangerous[i]));
			pos = strstr(path, dangerous[i]);
		}
	}
	
	return 1;
}

// Parse command message from header value
struct command_message* command_parse_header(const char *header_value)
{
	if (!header_value) {
		return NULL;
	}
	
	struct command_message *cmd_msg = calloc(1, sizeof(struct command_message));
	if (!cmd_msg) {
		return NULL;
	}
	
	// Set defaults
	cmd_msg->timeout_seconds = COMMAND_DEFAULT_TIMEOUT;
	cmd_msg->working_directory = strdup("/tmp");
	cmd_msg->user = strdup("root");
	
	// Parse JSON format header
	json_object *json_obj = json_tokener_parse(header_value);
	if (!json_obj) {
		// Try simple format: just the command
		cmd_msg->command = strdup(header_value);
		if (strlen(cmd_msg->command) > COMMAND_MAX_LENGTH) {
			command_message_free(cmd_msg);
			return NULL;
		}
		return cmd_msg;
	}
	
	// Parse JSON fields
	json_object *cmd_obj, *args_obj, *timeout_obj, *workdir_obj, *user_obj;
	
	if (json_object_object_get_ex(json_obj, "command", &cmd_obj)) {
		cmd_msg->command = strdup(json_object_get_string(cmd_obj));
	}
	
	if (json_object_object_get_ex(json_obj, "args", &args_obj)) {
		cmd_msg->args = strdup(json_object_get_string(args_obj));
	}
	
	if (json_object_object_get_ex(json_obj, "timeout", &timeout_obj)) {
		cmd_msg->timeout_seconds = json_object_get_int(timeout_obj);
		if (cmd_msg->timeout_seconds <= 0 || cmd_msg->timeout_seconds > 300) {
			cmd_msg->timeout_seconds = COMMAND_DEFAULT_TIMEOUT;
		}
	}
	
	if (json_object_object_get_ex(json_obj, "workdir", &workdir_obj)) {
		free(cmd_msg->working_directory);
		cmd_msg->working_directory = strdup(json_object_get_string(workdir_obj));
		command_sanitize_path(cmd_msg->working_directory);
	}
	
	if (json_object_object_get_ex(json_obj, "user", &user_obj)) {
		free(cmd_msg->user);
		cmd_msg->user = strdup(json_object_get_string(user_obj));
	}
	
	json_object_put(json_obj);
	
	return cmd_msg;
}

void command_message_free(struct command_message *cmd_msg)
{
	if (!cmd_msg) {
		return;
	}
	
	free(cmd_msg->command);
	free(cmd_msg->args);
	free(cmd_msg->working_directory);
	free(cmd_msg->user);
	free(cmd_msg);
}

// Execute command safely with timeout and capture output
struct command_result* command_execute_safe(struct command_message *cmd_msg)
{
	if (!cmd_msg || !cmd_msg->command) {
		return NULL;
	}
	
	// Validate command safety
	if (!command_validate_safe(cmd_msg->command)) {
		log_message(NAME, L_WARNING, "unsafe command rejected: %s\n", cmd_msg->command);
		return NULL;
	}
	
	struct command_result *result = calloc(1, sizeof(struct command_result));
	if (!result) {
		return NULL;
	}
	
	// Record start time
	clock_gettime(CLOCK_MONOTONIC, &result->start_time);
	
	int stdout_pipe[2], stderr_pipe[2];
	pid_t pid;
	
	// Create pipes for capturing output
	if (pipe(stdout_pipe) == -1 || pipe(stderr_pipe) == -1) {
		log_message(NAME, L_WARNING, "failed to create pipes: %s\n", strerror(errno));
		free(result);
		return NULL;
	}
	
	// Fork process
	pid = fork();
	if (pid == -1) {
		log_message(NAME, L_WARNING, "failed to fork: %s\n", strerror(errno));
		close(stdout_pipe[0]);
		close(stdout_pipe[1]);
		close(stderr_pipe[0]);
		close(stderr_pipe[1]);
		free(result);
		return NULL;
	}
	
	if (pid == 0) {
		// Child process
		close(stdout_pipe[0]);
		close(stderr_pipe[0]);
		
		// Redirect stdout and stderr
		dup2(stdout_pipe[1], STDOUT_FILENO);
		dup2(stderr_pipe[1], STDERR_FILENO);
		
		close(stdout_pipe[1]);
		close(stderr_pipe[1]);
		
		// Change working directory if specified
		if (cmd_msg->working_directory) {
			if (chdir(cmd_msg->working_directory) != 0) {
				fprintf(stderr, "Failed to change directory to %s: %s\n", 
					cmd_msg->working_directory, strerror(errno));
			}
		}
		
		// Build command string
		char full_command[COMMAND_MAX_LENGTH + COMMAND_MAX_ARGS + 10];
		if (cmd_msg->args) {
			snprintf(full_command, sizeof(full_command), "%s %s", cmd_msg->command, cmd_msg->args);
		} else {
			snprintf(full_command, sizeof(full_command), "%s", cmd_msg->command);
		}
		
		// Execute command
		execl("/bin/sh", "sh", "-c", full_command, (char *)NULL);
		
		// If we get here, exec failed
		fprintf(stderr, "Failed to execute command: %s\n", strerror(errno));
		exit(127);
	} else {
		// Parent process
		close(stdout_pipe[1]);
		close(stderr_pipe[1]);
		
		// Set up timeout using alarm
		signal(SIGALRM, SIG_DFL);
		alarm(cmd_msg->timeout_seconds);
		
		// Wait for child process
		int status;
		pid_t wait_result = waitpid(pid, &status, 0);
		
		// Cancel alarm
		alarm(0);
		
		// Record end time
		clock_gettime(CLOCK_MONOTONIC, &result->end_time);
		
		// Calculate execution time in milliseconds
		result->execution_time_ms = 
			(result->end_time.tv_sec - result->start_time.tv_sec) * 1000 +
			(result->end_time.tv_nsec - result->start_time.tv_nsec) / 1000000;
		
		if (wait_result == -1) {
			if (errno == EINTR) {
				// Timeout occurred
				kill(pid, SIGKILL);
				waitpid(pid, &status, 0);
				result->exit_code = -1;
				result->stderr_data = strdup("Command timed out");
			} else {
				result->exit_code = -1;
				result->stderr_data = strdup("Failed to wait for command");
			}
		} else {
			result->exit_code = WEXITSTATUS(status);
		}
		
		// Read stdout
		char buffer[4096];
		ssize_t bytes_read;
		size_t total_stdout = 0, total_stderr = 0;
		
		// Read stdout with size limit
		while ((bytes_read = read(stdout_pipe[0], buffer, sizeof(buffer))) > 0) {
			if (total_stdout + bytes_read > COMMAND_MAX_OUTPUT_SIZE) {
				break;
			}
			
			char *new_stdout = realloc(result->stdout_data, total_stdout + bytes_read + 1);
			if (!new_stdout) {
				break;
			}
			
			result->stdout_data = new_stdout;
			memcpy(result->stdout_data + total_stdout, buffer, bytes_read);
			total_stdout += bytes_read;
			result->stdout_data[total_stdout] = '\0';
		}
		
		// Read stderr with size limit
		while ((bytes_read = read(stderr_pipe[0], buffer, sizeof(buffer))) > 0) {
			if (total_stderr + bytes_read > COMMAND_MAX_OUTPUT_SIZE) {
				break;
			}
			
			char *new_stderr = realloc(result->stderr_data, total_stderr + bytes_read + 1);
			if (!new_stderr) {
				break;
			}
			
			result->stderr_data = new_stderr;
			memcpy(result->stderr_data + total_stderr, buffer, bytes_read);
			total_stderr += bytes_read;
			result->stderr_data[total_stderr] = '\0';
		}
		
		close(stdout_pipe[0]);
		close(stderr_pipe[0]);
	}
	
	// Ensure we have empty strings instead of NULL
	if (!result->stdout_data) {
		result->stdout_data = strdup("");
	}
	if (!result->stderr_data) {
		result->stderr_data = strdup("");
	}
	
	log_message(NAME, L_NOTICE, "command executed: %s, exit_code: %d, time: %ldms\n",
		cmd_msg->command, result->exit_code, result->execution_time_ms);
	
	return result;
}

void command_result_free(struct command_result *result)
{
	if (!result) {
		return;
	}
	
	free(result->stdout_data);
	free(result->stderr_data);
	free(result);
}

// Convert command result to JSON response
char* command_result_to_json(struct command_result *result)
{
	if (!result) {
		return NULL;
	}
	
	json_object *json_response = json_object_new_object();
	json_object *json_status = json_object_new_string("success");
	json_object *json_exit_code = json_object_new_int(result->exit_code);
	json_object *json_stdout = json_object_new_string(result->stdout_data ? result->stdout_data : "");
	json_object *json_stderr = json_object_new_string(result->stderr_data ? result->stderr_data : "");
	json_object *json_exec_time = json_object_new_int64(result->execution_time_ms);
	
	// Add timestamp information
	char start_time_str[64], end_time_str[64];
	struct tm *tm_info;
	
	time_t start_time_t = result->start_time.tv_sec;
	time_t end_time_t = result->end_time.tv_sec;
	
	tm_info = localtime(&start_time_t);
	strftime(start_time_str, sizeof(start_time_str), "%Y-%m-%dT%H:%M:%S", tm_info);
	snprintf(start_time_str + strlen(start_time_str), 
		sizeof(start_time_str) - strlen(start_time_str), 
		".%03ld", result->start_time.tv_nsec / 1000000);
	
	tm_info = localtime(&end_time_t);
	strftime(end_time_str, sizeof(end_time_str), "%Y-%m-%dT%H:%M:%S", tm_info);
	snprintf(end_time_str + strlen(end_time_str), 
		sizeof(end_time_str) - strlen(end_time_str), 
		".%03ld", result->end_time.tv_nsec / 1000000);
	
	json_object *json_start_time = json_object_new_string(start_time_str);
	json_object *json_end_time = json_object_new_string(end_time_str);
	
	// Build response object
	json_object_object_add(json_response, "status", json_status);
	json_object_object_add(json_response, "exit_code", json_exit_code);
	json_object_object_add(json_response, "stdout", json_stdout);
	json_object_object_add(json_response, "stderr", json_stderr);
	json_object_object_add(json_response, "execution_time_ms", json_exec_time);
	json_object_object_add(json_response, "start_time", json_start_time);
	json_object_object_add(json_response, "end_time", json_end_time);
	
	const char *json_string = json_object_to_json_string(json_response);
	char *result_string = strdup(json_string);
	
	json_object_put(json_response);
	
	return result_string;
}
