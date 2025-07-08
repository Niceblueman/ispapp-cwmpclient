

#ifndef _ispappcwmp_JSON_H__
#define _ispappcwmp_JSON_H__

#ifdef JSONC
 #include <json-c/json.h>
#else
 #include <json/json.h>
#endif

int json_handle_get_parameter_value(char *line);
int json_handle_get_parameter_name(char *line);
int json_handle_get_parameter_attribute(char *line);
int json_handle_method_status(char *line);
int json_handle_set_parameter(char *line);
int json_handle_deviceid(char *line);
int json_handle_add_object(char *line);
int json_handle_check_parameter_value_change(char *line);

#endif
