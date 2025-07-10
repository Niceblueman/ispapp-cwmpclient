#ifndef _ISPAPPCWMP_BACKUP_H__
#define _ISPAPPCWMP_BACKUP_H__

#include <libxml/tree.h>
#define BACKUP_DIR "/etc/easycwmp"
#define BACKUP_FILE BACKUP_DIR"/.backup.xml"

int backup_extract_transfer_complete(xmlNodePtr node, char **msg_out, int *method_id);
int backup_remove_transfer_complete(xmlNodePtr node);
int backup_update_fault_transfer_complete(xmlNodePtr node, int fault_code);
int backup_update_complete_time_transfer_complete(xmlNodePtr node);
int backup_load_event(void);
int backup_remove_event(xmlNodePtr b);
int backup_load_download(void);
int backup_load_upload(void);
int backup_remove_download(xmlNodePtr node);
int backup_remove_upload(xmlNodePtr node);
int backup_save_file(void);
void backup_load(void);
void backup_init(void);
void backup_add_acsurl(char *acs_url);
void backup_check_acs_url(void);
void backup_check_software_version(void);
xmlNodePtr backup_check_transfer_complete(void);
xmlNodePtr backup_tree_init(void);
xmlNodePtr backup_add_transfer_complete(char *command_key, int fault_code, char *start_time, int method_id);
xmlNodePtr backup_add_event(int code, char *key, int method_id);
xmlNodePtr backup_add_download(char *key, int delay, char *file_size, char *download_url, char *file_type, char *username, char *password);
xmlNodePtr backup_add_upload(char *key, int delay, char *upload_url, char *file_type, char *username, char *password);
int backup_update_all_complete_time_transfer_complete(void);
void backup_cleanup(void);
#endif