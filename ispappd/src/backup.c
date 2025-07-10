#define _GNU_SOURCE
#include <unistd.h>
#include <sys/stat.h>
#include <string.h>
#include <stdlib.h>
#include <uci.h>
#include <roxml.h>
#include "backup.h"
#include "config.h"
#include "xml.h"
#include "cwmp.h"
#include "messages.h"
#include "time.h"

xml_node_t *backup_tree = NULL;

#ifdef NO_XML
// Stub implementations when no XML library is available

void backup_init(void) { return; }
xml_node_t *backup_tree_init(void) { return NULL; }
int backup_save_file(void) { return 0; }
void backup_add_acsurl(char *acs_url) { return; }
void backup_check_acs_url(void) { return; }
void backup_check_software_version(void) { return; }
xml_node_t *backup_add_transfer_complete(char *command_key, int fault_code, char *start_time, int method_id) { return NULL; }
int backup_update_fault_transfer_complete(xml_node_t *node, int fault_code) { return 0; }
int backup_update_complete_time_transfer_complete(xml_node_t *node) { return 0; }
int backup_update_all_complete_time_transfer_complete(void) { return 0; }
int backup_extract_transfer_complete(xml_node_t *node, char **msg_out, int *method_id) { return 0; }
int backup_remove_transfer_complete(xml_node_t *node) { return 0; }
int backup_load_download(void) { return 0; }
int backup_load_upload(void) { return 0; }
int backup_remove_download(xml_node_t *node) { return 0; }
int backup_remove_upload(xml_node_t *node) { return 0; }
int backup_load_event(void) { return 0; }
int backup_remove_event(xml_node_t *node) { return 0; }
void str_replace_newline_byspace(char *str) { 
	// Keep this function even in NO_XML mode as it might be used elsewhere
	while(*str) {
		if(*str == '\n' || *str == '\r')
			*str = ' ';
		str++;
	}
}

#else
// Real implementations when XML library is available

// Use roxml node_t for all XML node pointers

typedef node_t xml_node_t;

void str_replace_newline_byspace(char *str)
{
	while(*str)
	{
		if(*str == '\n' || *str == '\r')
			*str = ' ';
		str++;
	}
}

void backup_init(void)
{
#ifdef NO_XML
	// No XML library available - backup functionality disabled
	return;
#elif defined(HAVE_LIBROXML)
#ifdef BACKUP_DATA_IN_CONFIG
	char *val;
	ispappcwmp_uci_init();
	ispappcwmp_uci_set_value("ispappcwmp", "backup", NULL, "backup");
	ispappcwmp_uci_commit();
	val = ispappcwmp_uci_get_value("ispappcwmp", "backup","data");
	if(val[0]=='\0') {
		ispappcwmp_uci_fini();
		return;
	}
	backup_tree = roxml_load_buf(val);
	ispappcwmp_uci_fini();
#else
	FILE *fp;

	if (access(BACKUP_DIR, F_OK) == -1 ) {
		mkdir(BACKUP_DIR, 0777);
	}
	if (access(BACKUP_FILE, F_OK) == -1 ) {
		return;
	}
	fp = fopen(BACKUP_FILE, "r");
	if (fp!=NULL) {
		backup_tree = roxml_load_doc(BACKUP_FILE);
		fclose(fp);
	}
#endif
	backup_load_download();
	backup_load_upload();
	backup_load_event();
	backup_update_all_complete_time_transfer_complete();
#elif defined(HAVE_MXML)
	// TODO: Implement mxml support or disable backup functionality
	return;
#endif
}

xml_node_t *backup_tree_init(void)
{
#ifdef NO_XML
	return NULL;
#elif defined(HAVE_LIBROXML)
	xml_node_t *xml;

	backup_tree = roxml_load_buf("<backup_file/>");
	if (!backup_tree) return NULL;
	xml = roxml_add_node(backup_tree, 0, ROXML_ELM_NODE, "cwmp", NULL);
	if (!xml) return NULL;
	return xml;
#elif defined(HAVE_MXML)
	// TODO: Implement mxml support
	return NULL;
#else
	return NULL;
#endif
}

int backup_save_file(void) {
#ifdef NO_XML
	return 0;
#elif defined(HAVE_LIBROXML)
#ifdef BACKUP_DATA_IN_CONFIG
	char *val;
	int len;
	if (backup_tree == NULL)
		return 0;
	ispappcwmp_uci_init();
	roxml_commit_changes(backup_tree, NULL, &val, 0);
	len = strlen(val);
	if (len > 0 && val[len - 1] == '\n')
		val[len - 1]= '\0';
	str_replace_newline_byspace(val);
	ispappcwmp_uci_set_value("ispappcwmp", "backup", "data", val);
	ispappcwmp_uci_commit();
	ispappcwmp_uci_fini();
	free(val);
	return 0;
#else
	FILE *fp;
	char *val;
	int len;

	if (backup_tree == NULL)
		return 0;

	fp = fopen(BACKUP_FILE, "w");
	if (fp!=NULL) {
		roxml_commit_changes(backup_tree, NULL, &val, 0);
		len = strlen(val);
		if (len > 0 && val[len - 1] == '\n')
			val[len - 1]= '\0';
		str_replace_newline_byspace(val);
		fprintf(fp, "%s", val);
		fclose(fp);
		free(val);
		return 0;
	}
	return -1;
#endif
#elif defined(HAVE_MXML)
	return 0;
#else
	return 0;
#endif
}

void backup_add_acsurl(char *acs_url)
{
#ifdef NO_XML
	return;
#elif defined(HAVE_LIBROXML)
	xml_node_t *data, *b;

	cwmp_clean();
	roxml_close(backup_tree);
	b = backup_tree_init();
	data = roxml_add_node(b, 0, ROXML_ELM_NODE, "acs_url", NULL);
	data = roxml_add_node(data, 0, ROXML_TXT_NODE, NULL, acs_url);
	backup_save_file();
	cwmp_add_event(EVENT_BOOTSTRAP, NULL, 0, EVENT_BACKUP);
	cwmp_add_inform_timer();
#elif defined(HAVE_MXML)
	return;
#endif
}

void backup_check_acs_url(void)
{
#ifdef NO_XML
	return;
#elif defined(HAVE_LIBROXML)
	xml_node_t *b;

	b = roxml_get_chld(backup_tree, "acs_url", 0);
	if (!b || (roxml_get_txt(b, 0) && 
		strcmp(config->acs->url, roxml_get_content(roxml_get_txt(b, 0), NULL, 0, NULL)) != 0)) {
		backup_add_acsurl(config->acs->url);
	}
#elif defined(HAVE_MXML)
	return;
#endif
}

void backup_check_software_version(void)
{
#ifdef NO_XML
	return;
#elif defined(HAVE_LIBROXML)
	xml_node_t *data, *b;

	b = roxml_get_chld(backup_tree, "cwmp", 0);
	if (!b)
		b = backup_tree_init();

	data = roxml_get_chld(b, "software_version", 0);
	if (data) {
		char *content = roxml_get_content(roxml_get_txt(data, 0), NULL, 0, NULL);
		if (content && strcmp(config->device->software_version, content) != 0) {
			cwmp_add_event(EVENT_VALUE_CHANGE, NULL, 0, EVENT_NO_BACKUP);
		}
		roxml_del_node(data);
	}
	data = roxml_add_node(b, 0, ROXML_ELM_NODE, "software_version", NULL);
	data = roxml_add_node(data, 0, ROXML_TXT_NODE, NULL, config->device->software_version);
	backup_save_file();
	cwmp_add_inform_timer();
#elif defined(HAVE_MXML)
	return;
#endif
}

xml_node_t *backup_add_transfer_complete(char *command_key, int fault_code, char *start_time, int method_id)
{
	xml_node_t  *data, *m, *b;
	char c[16];

	data = roxml_get_chld(backup_tree, "cwmp", 0);
	if (!data) return NULL;

	m = roxml_add_node(data, 0, ROXML_ELM_NODE, "transfer_complete", NULL);
	if (!m) return NULL;
	b = roxml_add_node(m, 0, ROXML_ELM_NODE, "command_key", NULL);
	if (!b) return NULL;
	b = roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, command_key);
	if (!b) return NULL;
	b = roxml_add_node(m, 0, ROXML_ELM_NODE, "fault_code", NULL);
	if (!b) return NULL;
	b = roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, fault_array[fault_code].code);
	if (!b) return NULL;
	b = roxml_add_node(m, 0, ROXML_ELM_NODE, "fault_string", NULL);
	if (!b) return NULL;
	b = roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, fault_array[fault_code].string);
	if (!b) return NULL;
	b = roxml_add_node(m, 0, ROXML_ELM_NODE, "start_time", NULL);
	if (!b) return NULL;
	b = roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, start_time);
	if (!b) return NULL;
	b = roxml_add_node(m, 0, ROXML_ELM_NODE, "complete_time", NULL);
	if (!b) return NULL;
	b = roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, UNKNOWN_TIME);
	if (!b) return NULL;
	b = roxml_add_node(m, 0, ROXML_ELM_NODE, "method_id", NULL);
	if (!b) return NULL;
	snprintf(c, sizeof(c), "%d", method_id);
	b = roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, c);
	if (!b) return NULL;

	backup_save_file();
	return m;
}

int backup_update_fault_transfer_complete(xml_node_t *node, int fault_code)
{
	xml_node_t  *b, *txt;

	b = roxml_get_chld(node, "fault_code", 0);
	if (!b) return -1;
	txt = roxml_get_txt(b, 0);
	if (txt) {
		roxml_del_node(txt);
	}
	b = roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, fault_array[fault_code].code);
	if (!b) return -1;

	b = roxml_get_chld(node, "fault_string", 0);
	if (!b) return -1;
	txt = roxml_get_txt(b, 0);
	if (txt) {
		roxml_del_node(txt);
	}
	b = roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, fault_array[fault_code].string);
	if (!b) return -1;

	backup_save_file();
	return 0;
}

int backup_update_complete_time_transfer_complete(xml_node_t *node)
{
	xml_node_t  *b, *txt;

	b = roxml_get_chld(node, "complete_time", 0);
	if (!b) return -1;
	txt = roxml_get_txt(b, 0);
	if (txt) {
		roxml_del_node(txt);
	}
	b = roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, mix_get_time());
	if (!b) return -1;

	backup_save_file();
	return 0;
}

int backup_update_all_complete_time_transfer_complete(void)
{
	int count, i;
	xml_node_t **nodes = roxml_xpath(backup_tree, ".//transfer_complete", &count);
	
	for (i = 0; i < count; i++) {
		xml_node_t *n = nodes[i];
		xml_node_t *b = roxml_get_chld(n, "complete_time", 0);
		if (!b) {
			roxml_release(nodes);
			return -1;
		}
		
		xml_node_t *txt = roxml_get_txt(b, 0);
		if (txt) {
			char *content = roxml_get_content(txt, NULL, 0, NULL);
			if (content && strcmp(content, UNKNOWN_TIME) != 0) continue;
			roxml_del_node(txt);
			roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, mix_get_time());
		}
	}
	
	roxml_release(nodes);
	backup_save_file();
	return 0;
}

xml_node_t *backup_check_transfer_complete(void)
{
	xml_node_t *data;
	data = roxml_get_chld(backup_tree, "transfer_complete", 0);
	return data;
}

int backup_extract_transfer_complete( xml_node_t *node, char **msg_out, int *method_id)
{
	xml_node_t *tree_m, *n;
	xml_node_t *b, *txt;
	char *val;

	tree_m = mxmlLoadString(NULL, CWMP_TRANSFER_COMPLETE_MESSAGE, MXML_OPAQUE_CALLBACK);
	if (!tree_m) goto error;

	if(xml_add_cwmpid(tree_m)) goto error;

	// Extract command_key using roxml
	b = roxml_get_chld(node, "command_key", 0);
	if (!b) goto error;
	n = mxmlFindElement(tree_m, tree_m, "CommandKey", NULL, NULL, MXML_DESCEND);
	if (!n) goto error;
	txt = roxml_get_txt(b, 0);
	if (txt) {
		val = roxml_get_content(txt, NULL, 0, NULL);
		n = mxmlNewOpaque(n, val ? val : "");
	} else {
		n = mxmlNewOpaque(n, "");
	}
	if (!n) goto error;

	// Extract fault_code using roxml
	b = roxml_get_chld(node, "fault_code", 0);
	if (!b) goto error;
	n = mxmlFindElement(tree_m, tree_m, "FaultCode", NULL, NULL, MXML_DESCEND);
	if (!n) goto error;
	txt = roxml_get_txt(b, 0);
	if (txt) {
		val = roxml_get_content(txt, NULL, 0, NULL);
		n = mxmlNewOpaque(n, val ? val : "0");
	} else {
		n = mxmlNewOpaque(n, "0");
	}
	if (!n) goto error;

	// Extract fault_string using roxml
	b = roxml_get_chld(node, "fault_string", 0);
	if (!b) goto error;
	n = mxmlFindElement(tree_m, tree_m, "FaultString", NULL, NULL, MXML_DESCEND);
	if (!n) goto error;
	txt = roxml_get_txt(b, 0);
	if (txt) {
		val = roxml_get_content(txt, NULL, 0, NULL);
		n = mxmlNewOpaque(n, val ? val : "");
	} else {
		n = mxmlNewOpaque(n, "");
	}
	if (!n) goto error;

	// Extract start_time using roxml
	b = roxml_get_chld(node, "start_time", 0);
	if (!b) goto error;
	n = mxmlFindElement(tree_m, tree_m, "StartTime", NULL, NULL, MXML_DESCEND);
	if (!n) goto error;
	txt = roxml_get_txt(b, 0);
	if (txt) {
		val = roxml_get_content(txt, NULL, 0, NULL);
		n = mxmlNewOpaque(n, val ? val : "");
	} else {
		n = mxmlNewOpaque(n, "");
	}
	if (!n) goto error;

	// Extract complete_time using roxml
	b = roxml_get_chld(node, "complete_time", 0);
	if (!b) goto error;
	n = mxmlFindElement(tree_m, tree_m, "CompleteTime", NULL, NULL, MXML_DESCEND);
	if (!n) goto error;
	txt = roxml_get_txt(b, 0);
	if (txt) {
		val = roxml_get_content(txt, NULL, 0, NULL);
		n = mxmlNewOpaque(n, val ? val : "");
	} else {
		n = mxmlNewOpaque(n, "");
	}
	if (!n) goto error;

	// Extract method_id using roxml
	b = roxml_get_chld(node, "method_id", 0);
	if (!b) goto error;
	txt = roxml_get_txt(b, 0);
	if (txt) {
		val = roxml_get_content(txt, NULL, 0, NULL);
		*method_id = val ? atoi(val) : 0;
	} else {
		*method_id = 0;
	}

	*msg_out = mxmlSaveAllocString(tree_m, xml_format_cb);
	mxmlDelete(tree_m);
	return 0;
error:
	mxmlDelete(tree_m);
	return -1;
}

int backup_remove_transfer_complete(xml_node_t *node)
{
	mxmlDelete(node);
	backup_save_file();
	return 0;
}

xml_node_t *backup_add_download(char *key, int delay, char *file_size, char *download_url, char *file_type, char *username, char *password)
{
	xml_node_t *data, *b, *n;
	char time_execute[16];

	if (snprintf(time_execute,sizeof(time_execute),"%u",delay + (unsigned int)time(NULL)) < 0) return NULL;

	data = roxml_get_chld(backup_tree, "cwmp", 0);
	if (!data) return NULL;
	b = roxml_add_node(data, 0, ROXML_ELM_NODE, "download", NULL);
	if (!b) return NULL;

	n = roxml_add_node(b, 0, ROXML_ELM_NODE, "command_key", NULL);
	if (!n) return NULL;
	n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, key);
	if (!n) return NULL;

	n = roxml_add_node(b, 0, ROXML_ELM_NODE, "file_type", NULL);
	if (!n) return NULL;
	n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, file_type);
	if (!n) return NULL;

	n = roxml_add_node(b, 0, ROXML_ELM_NODE, "url", NULL);
	if (!n) return NULL;
	n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, download_url);
	if (!n) return NULL;

	n = roxml_add_node(b, 0, ROXML_ELM_NODE, "username", NULL);
	if (!n) return NULL;
	n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, username);
	if (!n) return NULL;

	n = roxml_add_node(b, 0, ROXML_ELM_NODE, "password", NULL);
	if (!n) return NULL;
	n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, password);
	if (!n) return NULL;

	n = roxml_add_node(b, 0, ROXML_ELM_NODE, "file_size", NULL);
	if (!n) return NULL;
	n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, file_size);
	if (!n) return NULL;

	n = roxml_add_node(b, 0, ROXML_ELM_NODE, "time_execute", NULL);
	if (!n) return NULL;
	n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, time_execute);
	if (!n) return NULL;

	backup_save_file();
	return b;
}

xml_node_t *backup_add_upload(char *key, int delay, char *upload_url, char *file_type, char *username, char *password)
{
    xml_node_t *data, *b, *n;
    char time_execute[16];

    if (snprintf(time_execute,sizeof(time_execute),"%u",delay + (unsigned int)time(NULL)) < 0) return NULL;

    data = roxml_get_chld(backup_tree, "cwmp", 0);
    if (!data) return NULL;
    b = roxml_add_node(data, 0, ROXML_ELM_NODE, "upload", NULL);
    if (!b) return NULL;

    n = roxml_add_node(b, 0, ROXML_ELM_NODE, "command_key", NULL);
    if (!n) return NULL;
    n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, key);
    if (!n) return NULL;

    n = roxml_add_node(b, 0, ROXML_ELM_NODE, "file_type", NULL);
    if (!n) return NULL;
    n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, file_type);
    if (!n) return NULL;

    n = roxml_add_node(b, 0, ROXML_ELM_NODE, "url", NULL);
    if (!n) return NULL;
    n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, upload_url);
    if (!n) return NULL;

    n = roxml_add_node(b, 0, ROXML_ELM_NODE, "username", NULL);
    if (!n) return NULL;
    n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, username);
    if (!n) return NULL;

    n = roxml_add_node(b, 0, ROXML_ELM_NODE, "password", NULL);
    if (!n) return NULL;
    n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, password);
    if (!n) return NULL;

    n = roxml_add_node(b, 0, ROXML_ELM_NODE, "time_execute", NULL);
    if (!n) return NULL;
    n = roxml_add_node(n, 0, ROXML_TXT_NODE, NULL, time_execute);
    if (!n) return NULL;

    backup_save_file();
    return b;
}

int backup_load_download(void)
{
    int delay = 0;
    unsigned int t;
    xml_node_t *data, *b, *c;
    char *download_url = NULL, *file_size = NULL,
        *command_key = NULL, *file_type = NULL,
        *username = NULL, *password = NULL;

    data = roxml_get_chld(backup_tree, "cwmp", 0);
    if (!data) return -1;
    int count = roxml_get_chld_nb(data);
    for (int i = 0; i < count; i++) {
        b = roxml_get_chld(data, NULL, i);
        if (!b) continue;
        char *node_name = roxml_get_name(b, NULL, 0);
        if (strcmp(node_name, "download") != 0) continue;

        c = roxml_get_chld(b, "command_key", 0);
        if (c && roxml_get_txt(c, 0))
            command_key = roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL);
        else
            command_key = "";

        c = roxml_get_chld(b, "url", 0);
        if (c && roxml_get_txt(c, 0))
            download_url = roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL);
        else
            download_url = "";

        c = roxml_get_chld(b, "username", 0);
        if (c && roxml_get_txt(c, 0))
            username = roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL);
        else
            username = "";

        c = roxml_get_chld(b, "password", 0);
        if (c && roxml_get_txt(c, 0))
            password = roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL);
        else
            password = "";

        c = roxml_get_chld(b, "file_size", 0);
        if (c && roxml_get_txt(c, 0))
            file_size = roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL);
        else
            file_size = "";

        c = roxml_get_chld(b, "time_execute", 0);
        if (c && roxml_get_txt(c, 0)) {
            sscanf(roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL), "%u", &t);
            delay = t - time(NULL);
        } else {
            delay = 0;
        }

        c = roxml_get_chld(b, "file_type", 0);
        if (c && roxml_get_txt(c, 0))
            file_type = roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL);
        else
            file_type = "";

        cwmp_add_download(command_key, delay, file_size, download_url, file_type, username, password, b);
    }
    return 0;
}

int backup_load_upload(void)
{
    int delay = 0;
    unsigned int t;
    xml_node_t *data, *b, *c;
    char *upload_url = NULL,
        *command_key = NULL, *file_type = NULL,
        *username = NULL, *password = NULL;

    data = roxml_get_chld(backup_tree, "cwmp", 0);
    if (!data) return -1;
    int count = roxml_get_chld_nb(data);
    for (int i = 0; i < count; i++) {
        b = roxml_get_chld(data, NULL, i);
        if (!b) continue;
        char *node_name = roxml_get_name(b, NULL, 0);
        if (strcmp(node_name, "upload") != 0) continue;

        c = roxml_get_chld(b, "command_key", 0);
        if (c && roxml_get_txt(c, 0))
            command_key = roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL);
        else
            command_key = "";

        c = roxml_get_chld(b, "url", 0);
        if (c && roxml_get_txt(c, 0))
            upload_url = roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL);
        else
            upload_url = "";

        c = roxml_get_chld(b, "username", 0);
        if (c && roxml_get_txt(c, 0))
            username = roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL);
        else
            username = "";

        c = roxml_get_chld(b, "password", 0);
        if (c && roxml_get_txt(c, 0))
            password = roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL);
        else
            password = "";

        c = roxml_get_chld(b, "time_execute", 0);
        if (c && roxml_get_txt(c, 0)) {
            sscanf(roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL), "%u", &t);
            delay = t - time(NULL);
        } else {
            delay = 0;
        }

        c = roxml_get_chld(b, "file_type", 0);
        if (c && roxml_get_txt(c, 0))
            file_type = roxml_get_content(roxml_get_txt(c, 0), NULL, 0, NULL);
        else
            file_type = "";

        cwmp_add_upload(command_key, delay, upload_url, file_type, username, password, b);
    }
    return 0;
}

int backup_remove_download(xml_node_t *node)
{
	mxmlDelete(node);
	backup_save_file();
	return 0;
}

int backup_remove_upload(xml_node_t *node)
{
	mxmlDelete(node);
	backup_save_file();
	return 0;
}

xml_node_t *backup_add_event(int code, char *key, int method_id)
{
	xml_node_t *b = backup_tree, *n, *data;
	char *e = NULL, *c = NULL;

	data = roxml_get_chld(backup_tree, "cwmp", 0);
	if (!data) goto error;
	n = roxml_add_node(data, 0, ROXML_ELM_NODE, "event", NULL);
	if (!n) goto error;

	if (asprintf(&e, "%d", code) == -1) goto error;
	b = roxml_add_node(n, 0, ROXML_ELM_NODE, "event_number", NULL);
	if (!b) goto error;
	b = roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, e);
	if (!b) goto error;

	if(key) {
		b = roxml_add_node(n, 0, ROXML_ELM_NODE, "event_key", NULL);
		if (!b) goto error;
		b = roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, key);
		if (!b) goto error;
	}

	if (method_id) {
		if (asprintf(&c, "%d", method_id) == -1) goto error;
		b = roxml_add_node(n, 0, ROXML_ELM_NODE, "event_method_id", NULL);
		if (!b) goto error;
		b = roxml_add_node(b, 0, ROXML_TXT_NODE, NULL, c);
		if (!b) goto error;
	}

	backup_save_file();

out:
	free(e);
	free(c);
	return n;

error:
	free(e);
	free(c);
	return NULL;
}

int backup_load_event(void)
{
	xml_node_t *data, *b, *c;
	char *event_num = NULL, *key = NULL;
	int method_id = 0;
	struct event *e;

	data = mxmlFindElement(backup_tree, backup_tree, "cwmp", NULL, NULL, MXML_DESCEND);
	if (!data) return -1;
	b = data;
	while (b = mxmlFindElement(b, data, "event", NULL, NULL, MXML_DESCEND)) {
		c = mxmlFindElement(b, b, "event_number",NULL, NULL, MXML_DESCEND);
		if (!c || !c->child || c->child->type != MXML_OPAQUE) return -1;
		event_num = c->child->value.opaque;

		c = mxmlFindElement(b, b, "event_key", NULL, NULL, MXML_DESCEND);
		if (c && c->child && c->child->type == MXML_OPAQUE && c->child->value.opaque) {
			c = c->child;
			key = xml_get_value_with_whitespace(&c, c->parent);
		}
		else
			key = NULL;

		c = mxmlFindElement(b, b, "event_method_id", NULL, NULL, MXML_DESCEND);
		if(c && c->child && c->child->type == MXML_OPAQUE)
			method_id = atoi(c->child->value.opaque);

		if(event_num) {
			if (e = cwmp_add_event(atoi(event_num), key, method_id, EVENT_NO_BACKUP))
				e->backup_node = b;
			cwmp_add_inform_timer();
		}
		FREE(key);
	}
	return 0;
}

int backup_remove_event(xml_node_t *b)
{
#ifdef HAVE_LIBROXML
	roxml_del_node(b);
#elif defined(HAVE_MXML)
	mxmlDelete(b);
#endif
	backup_save_file();
	return 0;
}

#endif // NO_XML
