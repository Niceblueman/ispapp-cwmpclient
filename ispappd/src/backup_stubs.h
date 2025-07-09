#ifndef _BACKUP_STUBS_H_
#define _BACKUP_STUBS_H_

#ifdef NO_XML
// Stub implementations when XML is not available

// Global variables stubs
extern void *backup_tree;

// Function stubs
static inline void backup_init(void) { return; }
static inline void *backup_tree_init(void) { return NULL; }
static inline int backup_save_file(void) { return 0; }
static inline void backup_add_acsurl(char *acs_url) { return; }
static inline void backup_check_acs_url(void) { return; }
static inline void backup_check_software_version(void) { return; }
static inline void *backup_add_transfer_complete(char *command_key, int fault_code, char *start_time, int method_id) { return NULL; }
static inline int backup_update_fault_transfer_complete(void *node, int fault_code) { return 0; }
static inline int backup_update_complete_time_transfer_complete(void *node) { return 0; }
static inline int backup_update_all_complete_time_transfer_complete(void) { return 0; }
static inline int backup_extract_transfer_complete(void *node, char **msg_out, int *method_id) { return 0; }
static inline int backup_remove_transfer_complete(void *node) { return 0; }
static inline int backup_load_download(void) { return 0; }
static inline int backup_load_upload(void) { return 0; }
static inline int backup_remove_download(void *node) { return 0; }
static inline int backup_remove_upload(void *node) { return 0; }
static inline int backup_load_event(void) { return 0; }
static inline int backup_remove_event(void *node) { return 0; }

#endif // NO_XML
#endif // _BACKUP_STUBS_H_
