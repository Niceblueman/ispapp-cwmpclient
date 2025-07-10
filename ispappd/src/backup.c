#include <unistd.h>
#include <sys/stat.h>
#include <uci.h>
#include <libxml2/libxml/parser.h>
#include <libxml2/libxml/tree.h>
#include "backup.h"
#include "config.h"
#include "xml.h"
#include "cwmp.h"
#include "messages.h"
#include "time.h"
#include "libxml_helpers.h"

xmlDocPtr backup_doc = NULL;
xmlNodePtr backup_tree = NULL;

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
	// Initialize libxml2
	xmlInitParser();
	
#ifdef BACKUP_DATA_IN_CONFIG
	char *val;
	easycwmp_uci_init();
	easycwmp_uci_set_value("easycwmp", "backup", NULL, "backup");
	easycwmp_uci_commit();
	val = easycwmp_uci_get_value("easycwmp", "backup","data");
	if(val[0]=='\0') {
		easycwmp_uci_fini();
		return;
	}
	backup_doc = xmlLoadStringDoc(val);
	if (backup_doc)
		backup_tree = xmlDocGetRootElement(backup_doc);
	easycwmp_uci_fini();
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
		fclose(fp);
		backup_doc = xmlParseFile(BACKUP_FILE);
		if (backup_doc)
			backup_tree = xmlDocGetRootElement(backup_doc);
	}
#endif
	backup_load_download();
	backup_load_upload();
	backup_load_event();
	backup_update_all_complete_time_transfer_complete();
}

xmlNodePtr backup_tree_init(void)
{
	xmlNodePtr xml;

	// Free any existing document
	if (backup_doc) {
		xmlFreeDoc(backup_doc);
		backup_doc = NULL;
		backup_tree = NULL;
	}
	
	// Create a new document
	backup_doc = xmlNewDoc((const xmlChar*)"1.0");
	if (!backup_doc) return NULL;
	
	// Create root element
	backup_tree = xmlNewNode(NULL, (const xmlChar*)"backup_file");
	xmlDocSetRootElement(backup_doc, backup_tree);
	if (!backup_tree) return NULL;
	
	// Create cwmp element
	xml = xmlNewChild(backup_tree, NULL, (const xmlChar*)"cwmp", NULL);
	if (!xml) return NULL;
	
	return xml;
}

int backup_save_file(void) {
#ifdef BACKUP_DATA_IN_CONFIG
	xmlChar *xmlbuff;
	int buffersize;
	char *val;
	
	if (backup_doc == NULL)
		return 0;
		
	xmlDocDumpMemory(backup_doc, &xmlbuff, &buffersize);
	val = strdup((char*)xmlbuff);
	xmlFree(xmlbuff);
	
	if (val) {
		int len = strlen(val);
		if (len > 0 && val[len - 1] == '\n')
			val[len - 1]= '\0';
		str_replace_newline_byspace(val);
		
		easycwmp_uci_init();
		easycwmp_uci_set_value("easycwmp", "backup", "data", val);
		easycwmp_uci_commit();
		easycwmp_uci_fini();
		free(val);
	}
	return 0;
#else
	FILE *fp;
	xmlChar *xmlbuff;
	int buffersize;
	char *val;

	if (backup_doc == NULL)
		return 0;

	fp = fopen(BACKUP_FILE, "w");
	if (fp!=NULL) {
		xmlDocDumpMemory(backup_doc, &xmlbuff, &buffersize);
		val = strdup((char*)xmlbuff);
		xmlFree(xmlbuff);
		
		if (val) {
			int len = strlen(val);
			if (len > 0 && val[len - 1] == '\n')
				val[len - 1]= '\0';
			str_replace_newline_byspace(val);
			fprintf(fp, "%s", val);
			free(val);
		}
		fclose(fp);
		return 0;
	}
	return -1;
#endif
}

void backup_add_acsurl(char *acs_url)
{
	xmlNodePtr data, *b;

	cwmp_clean();
	
	// Free existing document if any
	if (backup_doc) {
		xmlFreeDoc(backup_doc);
		backup_doc = NULL;
		backup_tree = NULL;
	}
	
	b = backup_tree_init();
	data = xmlNewChild(b, NULL, (const xmlChar*)"acs_url", NULL);
	xmlNewOpaque(data, acs_url);
	backup_save_file();
	cwmp_add_event(EVENT_BOOTSTRAP, NULL, 0, EVENT_BACKUP);
	cwmp_add_inform_timer();
}

void backup_check_acs_url(void)
{
	xmlNodePtr b;

	b = xmlFindElementByName(backup_tree, "acs_url");
	if (!b || (b->children && b->children->type == XML_TEXT_NODE && b->children->content &&
		strcmp(config->acs->url, (const char*)b->children->content) != 0)) {
		backup_add_acsurl(config->acs->url);
	}
}

void backup_check_software_version(void)
{
	xmlNodePtr data, b;

	b = xmlFindElementByName(backup_tree, "cwmp");
	if (!b)
		b = backup_tree_init();

	data = xmlFindElementByName(b, "software_version");
	if (data) {
		if (data->children && data->children->type == XML_TEXT_NODE && data->children->content &&
			strcmp(config->device->software_version, (const char*)data->children->content) != 0) {
			cwmp_add_event(EVENT_VALUE_CHANGE, NULL, 0, EVENT_NO_BACKUP);
		}
		xmlUnlinkNode(data);
		xmlFreeNode(data);
	}
	data = xmlNewChild(b, NULL, (const xmlChar*)"software_version", NULL);
	xmlNewOpaque(data, config->device->software_version);
	backup_save_file();
	cwmp_add_inform_timer();	
}

xmlNodePtr backup_add_transfer_complete(char *command_key, int fault_code, char *start_time, int method_id)
{
	xmlNodePtr data, m, b;
	char c[16];

	data = xmlFindElementByName(backup_tree, "cwmp");
	if (!data) return NULL;

	m = xmlNewChild(data, NULL, (const xmlChar*)"transfer_complete", NULL);
	if (!m) return NULL;
	
	b = xmlNewChild(m, NULL, (const xmlChar*)"command_key", NULL);
	if (!b) return NULL;
	xmlNewOpaque(b, command_key);
	
	b = xmlNewChild(m, NULL, (const xmlChar*)"fault_code", NULL);
	if (!b) return NULL;
	xmlNewOpaque(b, fault_array[fault_code].code);
	
	b = xmlNewChild(m, NULL, (const xmlChar*)"fault_string", NULL);
	if (!b) return NULL;
	xmlNewOpaque(b, fault_array[fault_code].string);
	
	b = xmlNewChild(m, NULL, (const xmlChar*)"start_time", NULL);
	if (!b) return NULL;
	xmlNewOpaque(b, start_time);
	
	b = xmlNewChild(m, NULL, (const xmlChar*)"complete_time", NULL);
	if (!b) return NULL;
	xmlNewOpaque(b, UNKNOWN_TIME);
	
	b = xmlNewChild(m, NULL, (const xmlChar*)"method_id", NULL);
	if (!b) return NULL;
	snprintf(c, sizeof(c), "%d", method_id);
	xmlNewOpaque(b, c);

	backup_save_file();
	return m;
}

int backup_update_fault_transfer_complete(xmlNodePtr node, int fault_code)
{
	xmlNodePtr b;

	b = xmlFindElementByName(node, "fault_code");
	if (!b) return -1;
	
	// Remove any existing content
	xmlNodeSetContent(b, NULL);
	// Add new content
	if (!xmlNewOpaque(b, fault_array[fault_code].code))
		return -1;

	b = xmlFindElementByName(node, "fault_string");
	if (!b) return -1;
	
	// Remove any existing content
	xmlNodeSetContent(b, NULL);
	// Add new content
	if (!xmlNewOpaque(b, fault_array[fault_code].string))
		return -1;

	backup_save_file();
	return 0;
}

int backup_update_complete_time_transfer_complete(xmlNodePtr node)
{
	xmlNodePtr b;

	b = xmlFindElementByName(node, "complete_time");
	if (!b) return -1;
	
	// Remove any existing content
	xmlNodeSetContent(b, NULL);
	// Add new content
	if (!xmlNewOpaque(b, mix_get_time()))
		return -1;

	backup_save_file();
	return 0;
}

int backup_update_all_complete_time_transfer_complete(void)
{
	xmlNodePtr b, n = backup_tree;
	xmlNodePtr current = NULL;
	
	if (!backup_tree) return 0;
	
	// Find first transfer_complete element
	n = xmlFindElementByName(backup_tree, "transfer_complete");
	
	while (n) {
		b = xmlFindElementByName(n, "complete_time");
		if (!b) return -1;
		
		if (b->children && b->children->type == XML_TEXT_NODE && b->children->content) {
			if (strcmp((const char*)b->children->content, UNKNOWN_TIME) != 0) {
				// Skip this one, find next
				current = n;
				n = NULL;
				// Search for the next transfer_complete starting from current
				for (xmlNodePtr tmp = xmlWalkNext(current); tmp; tmp = xmlWalkNext(tmp)) {
					if (tmp->type == XML_ELEMENT_NODE && !strcmp((const char*)tmp->name, "transfer_complete")) {
						n = tmp;
						break;
					}
				}
				continue;
			}
			
			// Remove any existing content
			xmlNodeSetContent(b, NULL);
			// Add new content
			if (!xmlNewOpaque(b, mix_get_time()))
				return -1;
		}
		
		// Move to next transfer_complete
		current = n;
		n = NULL;
		// Search for the next transfer_complete starting from current
		for (xmlNodePtr tmp = xmlWalkNext(current); tmp; tmp = xmlWalkNext(tmp)) {
			if (tmp->type == XML_ELEMENT_NODE && !strcmp((const char*)tmp->name, "transfer_complete")) {
				n = tmp;
				break;
			}
		}
	}
	
	backup_save_file();
	return 0;
}

xmlNodePtr backup_check_transfer_complete(void)
{
	xmlNodePtr data;
	data = xmlFindElementByName(backup_tree, "transfer_complete");
	return data;
}

int backup_extract_transfer_complete(xmlNodePtr node, char **msg_out, int *method_id)
{
	xmlDocPtr tree_doc;
	xmlNodePtr tree_m, b, n;
	char *val;

	// Parse the transfer complete message template
	tree_doc = xmlLoadStringDoc(CWMP_TRANSFER_COMPLETE_MESSAGE);
	if (!tree_doc) goto error;
	tree_m = xmlDocGetRootElement(tree_doc);
	if (!tree_m) goto error;

	if(xml_add_cwmpid(tree_doc)) goto error;

	b = xmlFindElementByName(node, "command_key");
	if (!b) goto error;
	n = xmlFindElementByName(tree_m, "CommandKey");
	if (!n) goto error;
	if (b->children && b->children->type == XML_TEXT_NODE && b->children->content) {
		val = xml_get_value_with_whitespace(b->children, b);
		xmlNewOpaque(n, val);
		FREE(val);
	}
	else {
		xmlNewOpaque(n, "");
	}

	b = xmlFindElementByName(node, "fault_code");
	if (!b || !b->children) goto error;
	n = xmlFindElementByName(tree_m, "FaultCode");
	if (!n) goto error;
	xmlNewOpaque(n, (const char*)b->children->content);

	b = xmlFindElementByName(node, "fault_string");
	if (!b) goto error;
	if (b->children && b->children->type == XML_TEXT_NODE && b->children->content) {
		n = xmlFindElementByName(tree_m, "FaultString");
		if (!n) goto error;
		char *c = xml_get_value_with_whitespace(b->children, b);
		xmlNewOpaque(n, c);
		free(c);
	}

	b = xmlFindElementByName(node, "start_time");
	if (!b || !b->children) goto error;
	n = xmlFindElementByName(tree_m, "StartTime");
	if (!n) goto error;
	xmlNewOpaque(n, (const char*)b->children->content);

	b = xmlFindElementByName(node, "complete_time");
	if (!b || !b->children) goto error;
	n = xmlFindElementByName(tree_m, "CompleteTime");
	if (!n) goto error;
	xmlNewOpaque(n, (const char*)b->children->content);

	b = xmlFindElementByName(node, "method_id");
	if (!b || !b->children) goto error;
	*method_id = atoi((const char*)b->children->content);

	// Save to string
	xmlChar *xmlbuff;
	int buffersize;
	xmlDocDumpMemoryEnc(tree_doc, &xmlbuff, &buffersize, "UTF-8");
	*msg_out = strdup((char*)xmlbuff);
	xmlFree(xmlbuff);
	
	xmlFreeDoc(tree_doc);
	return 0;
error:
	if (tree_doc) xmlFreeDoc(tree_doc);
	return -1;
}

int backup_remove_transfer_complete(xmlNodePtr node)
{
	xmlUnlinkNode(node);
	xmlFreeNode(node);
	backup_save_file();
	return 0;
}

xmlNodePtr backup_add_download(char *key, int delay, char *file_size, char *download_url, char *file_type, char *username, char *password)
{
	xmlNodePtr data, b, n;
	char time_execute[16];

	if (snprintf(time_execute,sizeof(time_execute),"%u",delay + (unsigned int)time(NULL)) < 0) return NULL;

	data = xmlFindElementByName(backup_tree, "cwmp");
	if (!data) return NULL;
	b = xmlNewChild(data, NULL, (const xmlChar*)"download", NULL);
	if (!b) return NULL;

	n = xmlNewChild(b, NULL, (const xmlChar*)"command_key", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, key);

	n = xmlNewChild(b, NULL, (const xmlChar*)"file_type", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, file_type);

	n = xmlNewChild(b, NULL, (const xmlChar*)"url", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, download_url);

	n = xmlNewChild(b, NULL, (const xmlChar*)"username", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, username);

	n = xmlNewChild(b, NULL, (const xmlChar*)"password", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, password);

	n = xmlNewChild(b, NULL, (const xmlChar*)"file_size", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, file_size);

	n = xmlNewChild(b, NULL, (const xmlChar*)"time_execute", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, time_execute);

	backup_save_file();
	return b;
}

xmlNodePtr backup_add_upload(char *key, int delay, char *upload_url, char *file_type, char *username, char *password)
{
	xmlNodePtr data, b, n;
	char time_execute[16];

	if (snprintf(time_execute,sizeof(time_execute),"%u",delay + (unsigned int)time(NULL)) < 0) return NULL;

	data = xmlFindElementByName(backup_tree, "cwmp");
	if (!data) return NULL;
	b = xmlNewChild(data, NULL, (const xmlChar*)"upload", NULL);
	if (!b) return NULL;

	n = xmlNewChild(b, NULL, (const xmlChar*)"command_key", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, key);

	n = xmlNewChild(b, NULL, (const xmlChar*)"file_type", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, file_type);

	n = xmlNewChild(b, NULL, (const xmlChar*)"url", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, upload_url);

	n = xmlNewChild(b, NULL, (const xmlChar*)"username", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, username);

	n = xmlNewChild(b, NULL, (const xmlChar*)"password", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, password);

	n = xmlNewChild(b, NULL, (const xmlChar*)"time_execute", NULL);
	if (!n) return NULL;
	xmlNewOpaque(n, time_execute);

	backup_save_file();
	return b;
}

int backup_load_download(void)
{
	int delay = 0;
	unsigned int t;
	xmlNodePtr data, b, c;
	char *download_url = NULL, *file_size = NULL,
		*command_key = NULL, *file_type = NULL,
		*username = NULL, *password = NULL, *val = NULL;

	data = xmlFindElementByName(backup_tree, "cwmp");
	if (!data) return -1;
	
	// Find first download element
	b = xmlFindElementByName(data, "download");
	
	while (b) {
		xmlNodePtr next_download = NULL;
		
		c = xmlFindElementByName(b, "command_key");
		if (!c) return -1;
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content) {
			val = xml_get_value_with_whitespace(c->children, c);
			command_key = val;
		}
		else
			command_key = strdup("");

		c = xmlFindElementByName(b, "url");
		if (!c) goto error;
		if (c->children && c->children->content)
			download_url = (char*)c->children->content;
		else
			download_url = "";

		c = xmlFindElementByName(b, "username");
		if (!c) goto error;
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content) {
			val = xml_get_value_with_whitespace(c->children, c);
			username = val;
		}
		else
			username = strdup("");

		c = xmlFindElementByName(b, "password");
		if (!c) goto error;
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content) {
			val = xml_get_value_with_whitespace(c->children, c);
			password = val;
		}
		else
			password = strdup("");

		c = xmlFindElementByName(b, "file_size");
		if (!c) goto error;
		if (c->children && c->children->content)
			file_size = (char*)c->children->content;
		else
			file_size = "";

		c = xmlFindElementByName(b, "time_execute");
		if (!c) goto error;
		if (c->children && c->children->content) {
			sscanf((const char*)c->children->content, "%u", &t);
			delay = t - time(NULL);
		}

		c = xmlFindElementByName(b, "file_type");
		if (!c) goto error;
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content) {
			val = xml_get_value_with_whitespace(c->children, c);
			file_type = val;
		}
		else
			file_type = strdup("");

		cwmp_add_download(command_key, delay, file_size, download_url, file_type, username, password, b);
		FREE(command_key);
		FREE(username);
		FREE(password);
		FREE(file_type);
		
		// Find next download element
		for (xmlNodePtr tmp = xmlWalkNext(b); tmp; tmp = xmlWalkNext(tmp)) {
			if (tmp->type == XML_ELEMENT_NODE && !strcmp((const char*)tmp->name, "download")) {
				next_download = tmp;
				break;
			}
		}
		b = next_download;
	}
	return 0;
error:
	FREE(command_key);
	FREE(username);
	FREE(password);
	FREE(file_type);
	return -1;
}

int backup_load_upload(void)
{
	int delay = 0;
	unsigned int t;
	xmlNodePtr data, b, c;
	char *upload_url = NULL,
		*command_key = NULL, *file_type = NULL,
		*username = NULL, *password = NULL, *val = NULL;

	data = xmlFindElementByName(backup_tree, "cwmp");
	if (!data) return -1;
	
	// Find first upload element
	b = xmlFindElementByName(data, "upload");
	
	while (b) {
		xmlNodePtr next_upload = NULL;
		
		c = xmlFindElementByName(b, "command_key");
		if (!c) return -1;
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content) {
			val = xml_get_value_with_whitespace(c->children, c);
			command_key = val;
		}
		else
			command_key = strdup("");

		c = xmlFindElementByName(b, "url");
		if (!c) goto error;
		if (c->children && c->children->content)
			upload_url = (char*)c->children->content;
		else
			upload_url = "";

		c = xmlFindElementByName(b, "username");
		if (!c) goto error;
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content) {
			val = xml_get_value_with_whitespace(c->children, c);
			username = val;
		}
		else
			username = strdup("");

		c = xmlFindElementByName(b, "password");
		if (!c) goto error;
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content) {
			val = xml_get_value_with_whitespace(c->children, c);
			password = val;
		}
		else
			password = strdup("");

		c = xmlFindElementByName(b, "time_execute");
		if (!c) goto error;
		if (c->children && c->children->content) {
			sscanf((const char*)c->children->content, "%u", &t);
			delay = t - time(NULL);
		}

		c = xmlFindElementByName(b, "file_type");
		if (!c) goto error;
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content) {
			val = xml_get_value_with_whitespace(c->children, c);
			file_type = val;
		}
		else
			file_type = strdup("");

		cwmp_add_upload(command_key, delay, upload_url, file_type, username, password, b);
		FREE(command_key);
		FREE(username);
		FREE(password);
		FREE(file_type);
		
		// Find next upload element
		for (xmlNodePtr tmp = xmlWalkNext(b); tmp; tmp = xmlWalkNext(tmp)) {
			if (tmp->type == XML_ELEMENT_NODE && !strcmp((const char*)tmp->name, "upload")) {
				next_upload = tmp;
				break;
			}
		}
		b = next_upload;
	}
	return 0;
error:
	FREE(command_key);
	FREE(username);
	FREE(password);
	FREE(file_type);
	return -1;
}

int backup_remove_download(xmlNodePtr node)
{
	xmlUnlinkNode(node);
	xmlFreeNode(node);
	backup_save_file();
	return 0;
}

int backup_remove_upload(xmlNodePtr node)
{
	xmlUnlinkNode(node);
	xmlFreeNode(node);
	backup_save_file();
	return 0;
}

xmlNodePtr backup_add_event(int code, char *key, int method_id)
{
	xmlNodePtr b, n, data;
	char *e = NULL, *c = NULL;

	data = xmlFindElementByName(backup_tree, "cwmp");
	if (!data) goto error;
	n = xmlNewChild(data, NULL, (const xmlChar*)"event", NULL);
	if (!n) goto error;

	if (asprintf(&e, "%d", code) == -1) goto error;
	b = xmlNewChild(n, NULL, (const xmlChar*)"event_number", NULL);
	if (!b) goto error;
	if (!xmlNewOpaque(b, e)) goto error;

	if(key) {
		b = xmlNewChild(n, NULL, (const xmlChar*)"event_key", NULL);
		if (!b) goto error;
		if (!xmlNewOpaque(b, key)) goto error;
	}

	if (method_id) {
		if (asprintf(&c, "%d", method_id) == -1) goto error;
		b = xmlNewChild(n, NULL, (const xmlChar*)"event_method_id", NULL);
		if (!b) goto error;
		if (!xmlNewOpaque(b, c)) goto error;
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
	xmlNodePtr data, b, c;
	char *event_num = NULL, *key = NULL;
	int method_id = 0;
	struct event *e;

	data = xmlFindElementByName(backup_tree, "cwmp");
	if (!data) return -1;
	
	// Find first event element
	b = xmlFindElementByName(data, "event");
	
	while (b) {
		xmlNodePtr next_event = NULL;
		
		c = xmlFindElementByName(b, "event_number");
		if (!c || !c->children || c->children->type != XML_TEXT_NODE) return -1;
		event_num = (char*)c->children->content;

		c = xmlFindElementByName(b, "event_key");
		if (c && c->children && c->children->type == XML_TEXT_NODE && c->children->content) {
			key = xml_get_value_with_whitespace(c->children, c);
		}
		else
			key = NULL;

		c = xmlFindElementByName(b, "event_method_id");
		if(c && c->children && c->children->type == XML_TEXT_NODE && c->children->content)
			method_id = atoi((const char*)c->children->content);

		if(event_num) {
			if (e = cwmp_add_event(atoi(event_num), key, method_id, EVENT_NO_BACKUP))
				e->backup_node = b;
			cwmp_add_inform_timer();
		}
		FREE(key);
		
		// Find next event element
		for (xmlNodePtr tmp = xmlWalkNext(b); tmp; tmp = xmlWalkNext(tmp)) {
			if (tmp->type == XML_ELEMENT_NODE && !strcmp((const char*)tmp->name, "event")) {
				next_event = tmp;
				break;
			}
		}
		b = next_event;
	}
	return 0;
}

int backup_remove_event(xmlNodePtr b)
{
	xmlUnlinkNode(b);
	xmlFreeNode(b);
	backup_save_file();
	return 0;
}

void backup_cleanup(void)
{
	if (backup_doc) {
		xmlFreeDoc(backup_doc);
		backup_doc = NULL;
		backup_tree = NULL;
	}
	xmlCleanupParser();
}