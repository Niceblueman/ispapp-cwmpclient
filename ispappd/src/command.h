#ifndef _ISPAPPCWMP_COMMAND_H__
#define _ISPAPPCWMP_COMMAND_H__


// Command execution result structure
#ifdef __USE_POSIX
#include <sys/time.h>
struct command_result
{
	char *stdout_data;
	char *stderr_data;
	int exit_code;
	struct timespec start_time;
	struct timespec end_time;
	long execution_time_ms;
};
#else
#include <time.h>
struct mtimespec
{
	time_t tv_sec;        /* seconds */
	long   tv_nsec;       /* nanoseconds */
};

struct command_result
{
	char *stdout_data;
	char *stderr_data;
	int exit_code;
	struct mtimespec start_time; // Changed to timeval for non-POSIX
	struct mtimespec end_time;   // Changed to timeval for non-POSIX
	long execution_time_ms;
};
#endif
// Command message format structure
struct command_message
{
	char *command;
	char *args;
	int timeout_seconds;
	char *working_directory;
	char *user;
};

// Function prototypes
int command_init(void);
void command_cleanup(void);

struct command_result *command_execute_safe(struct command_message *cmd_msg);
void command_result_free(struct command_result *result);

char *command_result_to_json(struct command_result *result);
struct command_message *command_parse_header(const char *header_value);
void command_message_free(struct command_message *cmd_msg);

// Safety and validation functions
int command_validate_safe(const char *command);
int command_sanitize_path(char *path);

#define COMMAND_MAX_LENGTH 1024
#define COMMAND_MAX_ARGS 2048
#define COMMAND_DEFAULT_TIMEOUT 30
#define COMMAND_MAX_OUTPUT_SIZE (1024 * 1024) // 1MB max output

// Safe command whitelist patterns
extern const char *safe_command_patterns[];

#endif
