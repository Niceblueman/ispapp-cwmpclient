

#ifndef _ispappcwmp_EXTERNAL_H__
#define _ispappcwmp_EXTERNAL_H__
#include <libubox/list.h>

static char *fc_script = "/usr/sbin/ispappcwmp";
extern struct list_head external_list_parameter;

#define EXTERNAL_PROMPT "ispappcwmp>"

/*
 * external_parameter structure is used to get data from external command when a parameter method is triggered
 * The (*data) is used as notification for GetParameterAttribute; as writable for GetParameterNames; as value for GetParameterValues
 */
struct external_parameter {
	struct list_head list;
	char *name;
	char *data;
	char *type;
	char *fault_code;
};

void external_set_param_resp_status (char *status);
void external_fetch_set_param_resp_status (char **status);
void external_method_resp_status (char *status, char *fault);
void external_fetch_method_resp_status (char **status, char **fault);
void external_add_obj_resp (char *status, char *instance, char *fault);
void external_fetch_add_obj_resp (char **status, char **instance, char **fault);
int external_action_parameter_execute(char *command, char *class, char *name, char *arg);
int external_action_simple_execute(char *command, char *class, char *arg);
int external_action_download_execute(char *url, char *file_type, char *file_size, char *user_name, char *password);
int external_action_upload_execute(char *url, char *file_type, char *user_name, char *password);
int external_action_handle (int (*json_handle)(char *));
int external_init();
void external_exit();

void external_add_list_paramameter(char *param_name, char *param_data, char *param_type, char *fault_code);
void external_free_list_parameter();
void external_parameter_delete(struct external_parameter *external_parameter);
#endif

