#include "config.h"
#include "backup.h"
#include "xml.h"
#include "cwmp.h"
#include "messages.h"
#include "dtime.h"
#include "libxml_helpers.h"
#include "ispappcwmp.h"

xmlDocPtr backup_doc = NULL;
xmlNodePtr backup_tree = NULL;

void str_replace_newline_byspace(char *str)
{
	while (*str)
	{
		if (*str == '\n' || *str == '\r')
			*str = ' ';
		str++;
	}
}

void backup_init(void)
{
	xmlInitParser();

#ifdef BACKUP_DATA_IN_CONFIG
	char *val;
	ISPAPPCWMP_uci_init();
	ISPAPPCWMP_uci_set_value("ispapp", "backup", NULL, "backup");
	ISPAPPCWMP_uci_commit();
	val = ISPAPPCWMP_uci_get_value("ispapp", "backup", "data");
	if (!val || val[0] == '\0')
	{
		log_message(NAME, L_CRIT, "No backup data in UCI configuration\n");
		ISPAPPCWMP_uci_fini();
		return;
	}
	backup_doc = xmlLoadStringDoc(val);
	if (!backup_doc)
	{
		log_message(NAME, L_CRIT, "Failed to parse UCI backup data: %s\n", val);
		ISPAPPCWMP_uci_fini();
		return;
	}
	backup_tree = xmlDocGetRootElement(backup_doc);
	if (!backup_tree)
	{
		log_message(NAME, L_CRIT, "No root element in UCI backup XML\n");
		xmlFreeDoc(backup_doc);
		backup_doc = NULL;
		ISPAPPCWMP_uci_fini();
		return;
	}
	if (!backup_tree->name || !backup_tree->type || strcmp((const char *)backup_tree->name, "backup_file") != 0)
	{
		log_message(NAME, L_CRIT, "Invalid root node %p (name: %s, type: %d)\n",
					backup_tree, backup_tree->name ? (const char *)backup_tree->name : "NULL", backup_tree->type);
		xmlFreeDoc(backup_doc);
		backup_doc = NULL;
		backup_tree = NULL;
		ISPAPPCWMP_uci_fini();
		return;
	}
	// Log XML content and children of root
	xmlChar *xmlbuff;
	int buffersize;
	xmlDocDumpMemory(backup_doc, &xmlbuff, &buffersize);
	log_message(NAME, L_CRIT, "XML content: %s\n", (char *)xmlbuff);
	xmlFree(xmlbuff);
	for (xmlNodePtr child = backup_tree->children; child; child = child->next)
	{
		log_message(NAME, L_DEBUG, "Root child: node %p (name: %s, type: %d)\n",
					child, child->name ? (const char *)child->name : "NULL", child->type);
		if (child->type == XML_ELEMENT_NODE && strcmp((const char *)child->name, "cwmp") == 0)
		{
			log_message(NAME, L_DEBUG, "Found <cwmp> node at %p\n", child);
		}
	}
	ISPAPPCWMP_uci_fini();
#else
	FILE *fp;
	if (access(BACKUP_DIR, F_OK) == -1)
	{
		mkdir(BACKUP_DIR, 0777);
	}
	if (access(BACKUP_FILE, F_OK) == -1)
	{
		log_message(NAME, L_CRIT, "Backup file %s does not exist\n", BACKUP_FILE);
		return;
	}
	fp = fopen(BACKUP_FILE, "r");
	if (!fp)
	{
		log_message(NAME, L_CRIT, "Failed to open backup file %s\n", BACKUP_FILE);
		return;
	}
	fclose(fp);
	backup_doc = xmlParseFile(BACKUP_FILE);
	if (!backup_doc)
	{
		log_message(NAME, L_CRIT, "Failed to parse backup file %s\n", BACKUP_FILE);
		return;
	}
	backup_tree = xmlDocGetRootElement(backup_doc);
	if (!backup_tree)
	{
		log_message(NAME, L_CRIT, "No root element in backup file %s\n", BACKUP_FILE);
		xmlFreeDoc(backup_doc);
		backup_doc = NULL;
		return;
	}
	if (!backup_tree->name || !backup_tree->type || strcmp((const char *)backup_tree->name, "backup_file") != 0)
	{
		log_message(NAME, L_CRIT, "Invalid root node %p (name: %s, type: %d)\n",
					backup_tree, backup_tree->name ? (const char *)backup_tree->name : "NULL", backup_tree->type);
		xmlFreeDoc(backup_doc);
		backup_doc = NULL;
		backup_tree = NULL;
		return;
	}
	// Log XML content and children of root
	xmlChar *xmlbuff;
	int buffersize;
	xmlDocDumpMemory(backup_doc, &xmlbuff, &buffersize);
	log_message(NAME, L_CRIT, "XML content: %s\n", (char *)xmlbuff);
	xmlFree(xmlbuff);
	for (xmlNodePtr child = backup_tree->children; child; child = child->next)
	{
		log_message(NAME, L_DEBUG, "Root child: node %p (name: %s, type: %d)\n",
					child, child->name ? (const char *)child->name : "NULL", child->type);
		if (child->type == XML_ELEMENT_NODE && strcmp((const char *)child->name, "cwmp") == 0)
		{
			log_message(NAME, L_DEBUG, "Found <cwmp> node at %p\n", child);
		}
	}
#endif
	if (backup_load_download() < 0)
	{
		log_message(NAME, L_CRIT, "Failed to load downloads\n");
	}
	if (backup_load_upload() < 0)
	{
		log_message(NAME, L_CRIT, "Failed to load uploads\n");
	}
	if (backup_load_event() < 0)
	{
		log_message(NAME, L_CRIT, "Failed to load events\n");
	}
	backup_update_all_complete_time_transfer_complete();
}

xmlNodePtr backup_tree_init(void)
{
	xmlNodePtr xml;

	// Free any existing document
	if (backup_doc)
	{
		xmlFreeDoc(backup_doc);
		backup_doc = NULL;
		backup_tree = NULL;
	}

	// Create a new document
	backup_doc = xmlNewDoc((const xmlChar *)"1.0");
	if (!backup_doc)
		return NULL;

	// Create root element
	backup_tree = xmlNewNode(NULL, (const xmlChar *)"backup_file");
	xmlDocSetRootElement(backup_doc, backup_tree);
	if (!backup_tree)
		return NULL;

	// Create cwmp element
	xml = xmlNewChild(backup_tree, NULL, (const xmlChar *)"cwmp", NULL);
	if (!xml)
		return NULL;

	return xml;
}

int backup_save_file(void)
{
#ifdef BACKUP_DATA_IN_CONFIG
	xmlChar *xmlbuff;
	int buffersize;
	char *val;

	if (backup_doc == NULL)
		return 0;

	xmlDocDumpMemory(backup_doc, &xmlbuff, &buffersize);
	val = strdup((char *)xmlbuff);
	xmlFree(xmlbuff);

	if (val)
	{
		int len = strlen(val);
		if (len > 0 && val[len - 1] == '\n')
			val[len - 1] = '\0';
		str_replace_newline_byspace(val);

		ISPAPPCWMP_uci_init();
		ISPAPPCWMP_uci_set_value("easycwmp", "backup", "data", val);
		ISPAPPCWMP_uci_commit();
		ISPAPPCWMP_uci_fini();
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
	if (fp != NULL)
	{
		xmlDocDumpMemory(backup_doc, &xmlbuff, &buffersize);
		val = strdup((char *)xmlbuff);
		xmlFree(xmlbuff);

		if (val)
		{
			int len = strlen(val);
			if (len > 0 && val[len - 1] == '\n')
				val[len - 1] = '\0';
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
	xmlNodePtr data, b;

	cwmp_clean();

	// Free existing document if any
	if (backup_doc)
	{
		xmlFreeDoc(backup_doc);
		backup_doc = NULL;
		backup_tree = NULL;
	}

	b = backup_tree_init();
	data = xmlNewChild(b, NULL, (const xmlChar *)"acs_url", NULL);
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
			   strcmp(config->acs->url, (const char *)b->children->content) != 0))
	{
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
	if (data)
	{
		if (data->children && data->children->type == XML_TEXT_NODE && data->children->content &&
			strcmp(config->device->software_version, (const char *)data->children->content) != 0)
		{
			cwmp_add_event(EVENT_VALUE_CHANGE, NULL, 0, EVENT_NO_BACKUP);
		}
		xmlUnlinkNode(data);
		xmlFreeNode(data);
	}
	data = xmlNewChild(b, NULL, (const xmlChar *)"software_version", NULL);
	xmlNewOpaque(data, config->device->software_version);
	backup_save_file();
	cwmp_add_inform_timer();
}

xmlNodePtr backup_add_transfer_complete(char *command_key, int fault_code, char *start_time, int method_id)
{
	xmlNodePtr data, m, b;
	char c[16];

	data = xmlFindElementByName(backup_tree, "cwmp");
	if (!data)
		return NULL;

	m = xmlNewChild(data, NULL, (const xmlChar *)"transfer_complete", NULL);
	if (!m)
		return NULL;

	b = xmlNewChild(m, NULL, (const xmlChar *)"command_key", NULL);
	if (!b)
		return NULL;
	xmlNewOpaque(b, command_key);

	b = xmlNewChild(m, NULL, (const xmlChar *)"fault_code", NULL);
	if (!b)
		return NULL;
	xmlNewOpaque(b, fault_array[fault_code].code);

	b = xmlNewChild(m, NULL, (const xmlChar *)"fault_string", NULL);
	if (!b)
		return NULL;
	xmlNewOpaque(b, fault_array[fault_code].string);

	b = xmlNewChild(m, NULL, (const xmlChar *)"start_time", NULL);
	if (!b)
		return NULL;
	xmlNewOpaque(b, start_time);

	b = xmlNewChild(m, NULL, (const xmlChar *)"complete_time", NULL);
	if (!b)
		return NULL;
	xmlNewOpaque(b, UNKNOWN_TIME);

	b = xmlNewChild(m, NULL, (const xmlChar *)"method_id", NULL);
	if (!b)
		return NULL;
	snprintf(c, sizeof(c), "%d", method_id);
	xmlNewOpaque(b, c);

	backup_save_file();
	return m;
}

int backup_update_fault_transfer_complete(xmlNodePtr node, int fault_code)
{
	xmlNodePtr b;

	b = xmlFindElementByName(node, "fault_code");
	if (!b)
		return -1;

	// Remove any existing content
	xmlNodeSetContent(b, NULL);
	// Add new content
	if (!xmlNewOpaque(b, fault_array[fault_code].code))
		return -1;

	b = xmlFindElementByName(node, "fault_string");
	if (!b)
		return -1;

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
	if (!b)
		return -1;

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

	if (!backup_tree)
		return 0;

	// Find first transfer_complete element
	n = xmlFindElementByName(backup_tree, "transfer_complete");

	while (n)
	{
		b = xmlFindElementByName(n, "complete_time");
		if (!b)
			return -1;

		if (b->children && b->children->type == XML_TEXT_NODE && b->children->content)
		{
			if (strcmp((const char *)b->children->content, UNKNOWN_TIME) != 0)
			{
				// Skip this one, find next
				current = n;
				n = NULL;
				// Search for the next transfer_complete starting from current
				for (xmlNodePtr tmp = xmlWalkNextOne(current); tmp; tmp = xmlWalkNextOne(tmp))
				{
					if (tmp->type == XML_ELEMENT_NODE && !strcmp((const char *)tmp->name, "transfer_complete"))
					{
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
		for (xmlNodePtr tmp = xmlWalkNextOne(current); tmp; tmp = xmlWalkNextOne(tmp))
		{
			if (tmp->type == XML_ELEMENT_NODE && !strcmp((const char *)tmp->name, "transfer_complete"))
			{
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
	if (!tree_doc)
		goto error;
	tree_m = xmlDocGetRootElement(tree_doc);
	if (!tree_m)
		goto error;

	if (xml_add_cwmpid(tree_m))
		goto error;

	b = xmlFindElementByName(node, "command_key");
	if (!b)
		goto error;
	n = xmlFindElementByName(tree_m, "CommandKey");
	if (!n)
		goto error;
	if (b->children && b->children->type == XML_TEXT_NODE && b->children->content)
	{
		xmlNodePtr child = b->children;
		val = xml_get_value_with_whitespace(&child, b);
		xmlNewOpaque(n, val);
		free(val);
	}
	else
	{
		xmlNewOpaque(n, "");
	}

	b = xmlFindElementByName(node, "fault_code");
	if (!b || !b->children)
		goto error;
	n = xmlFindElementByName(tree_m, "FaultCode");
	if (!n)
		goto error;
	xmlNewOpaque(n, (const char *)b->children->content);

	b = xmlFindElementByName(node, "fault_string");
	if (!b)
		goto error;
	if (b->children && b->children->type == XML_TEXT_NODE && b->children->content)
	{
		n = xmlFindElementByName(tree_m, "FaultString");
		if (!n)
			goto error;
		xmlNodePtr child = b->children;
		char *c = xml_get_value_with_whitespace(&child, b);
		xmlNewOpaque(n, c);
		free(c);
	}

	b = xmlFindElementByName(node, "start_time");
	if (!b || !b->children)
		goto error;
	n = xmlFindElementByName(tree_m, "StartTime");
	if (!n)
		goto error;
	xmlNewOpaque(n, (const char *)b->children->content);

	b = xmlFindElementByName(node, "complete_time");
	if (!b || !b->children)
		goto error;
	n = xmlFindElementByName(tree_m, "CompleteTime");
	if (!n)
		goto error;
	xmlNewOpaque(n, (const char *)b->children->content);

	b = xmlFindElementByName(node, "method_id");
	if (!b || !b->children)
		goto error;
	*method_id = atoi((const char *)b->children->content);

	// Save to string
	xmlChar *xmlbuff;
	int buffersize;
	xmlDocDumpMemoryEnc(tree_doc, &xmlbuff, &buffersize, "UTF-8");
	*msg_out = strdup((char *)xmlbuff);
	xmlFree(xmlbuff);

	xmlFreeDoc(tree_doc);
	return 0;
error:
	if (tree_doc)
		xmlFreeDoc(tree_doc);
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

	if (snprintf(time_execute, sizeof(time_execute), "%u", delay + (unsigned int)time(NULL)) < 0)
		return NULL;

	data = xmlFindElementByName(backup_tree, "cwmp");
	if (!data)
		return NULL;
	b = xmlNewChild(data, NULL, (const xmlChar *)"download", NULL);
	if (!b)
		return NULL;

	n = xmlNewChild(b, NULL, (const xmlChar *)"command_key", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, key);

	n = xmlNewChild(b, NULL, (const xmlChar *)"file_type", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, file_type);

	n = xmlNewChild(b, NULL, (const xmlChar *)"url", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, download_url);

	n = xmlNewChild(b, NULL, (const xmlChar *)"username", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, username);

	n = xmlNewChild(b, NULL, (const xmlChar *)"password", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, password);

	n = xmlNewChild(b, NULL, (const xmlChar *)"file_size", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, file_size);

	n = xmlNewChild(b, NULL, (const xmlChar *)"time_execute", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, time_execute);

	backup_save_file();
	return b;
}

xmlNodePtr backup_add_upload(char *key, int delay, char *upload_url, char *file_type, char *username, char *password)
{
	xmlNodePtr data, b, n;
	char time_execute[16];

	if (snprintf(time_execute, sizeof(time_execute), "%u", delay + (unsigned int)time(NULL)) < 0)
		return NULL;

	data = xmlFindElementByName(backup_tree, "cwmp");
	if (!data)
		return NULL;
	b = xmlNewChild(data, NULL, (const xmlChar *)"upload", NULL);
	if (!b)
		return NULL;

	n = xmlNewChild(b, NULL, (const xmlChar *)"command_key", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, key);

	n = xmlNewChild(b, NULL, (const xmlChar *)"file_type", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, file_type);

	n = xmlNewChild(b, NULL, (const xmlChar *)"url", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, upload_url);

	n = xmlNewChild(b, NULL, (const xmlChar *)"username", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, username);

	n = xmlNewChild(b, NULL, (const xmlChar *)"password", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, password);

	n = xmlNewChild(b, NULL, (const xmlChar *)"time_execute", NULL);
	if (!n)
		return NULL;
	xmlNewOpaque(n, time_execute);

	backup_save_file();
	return b;
}

int backup_load_download(void)
{
	int delay = 0;
	unsigned int t;
	xmlNodePtr data = NULL, b = NULL, c = NULL;
	char *download_url = NULL, *file_size = NULL,
		 *command_key = NULL, *file_type = NULL,
		 *username = NULL, *password = NULL, *val = NULL;

	// Validate backup_tree first
	if (!backup_tree)
	{
		log_message(NAME, L_CRIT, "backup_tree is NULL in backup_load_download\n");
		return -1;
	}

	// Validate backup_doc
	if (!backup_doc)
	{
		log_message(NAME, L_CRIT, "backup_doc is NULL in backup_load_download\n");
		return -1;
	}

	// Log XML content for debugging
	xmlChar *xmlbuff = NULL;
	int buffersize = 0;
	xmlDocDumpMemory(backup_doc, &xmlbuff, &buffersize);
	if (xmlbuff)
	{
		log_message(NAME, L_DEBUG, "XML content: %s\n", (char *)xmlbuff);
		xmlFree(xmlbuff);
	}

	// Find cwmp element with proper error handling
	data = xmlFindElementByName(backup_tree, "cwmp");
	if (!data)
	{
		log_message(NAME, L_NOTICE, "No <cwmp> element found in backup_tree\n");
		return 0; // Not finding cwmp is not necessarily an error
	}

	// Validate the found node more carefully
	log_message(NAME, L_DEBUG, "Found <cwmp> node at %p\n", (void*)data);

	// IMPORTANT: Check basic pointer validity before accessing structure
	if (data == NULL)
	{
		log_message(NAME, L_CRIT, "<cwmp> node is NULL\n");
		return -1;
	}

	// Validate node structure - check type field first as it's least likely to cause issues
	if (data->type != XML_ELEMENT_NODE)
	{
		log_message(NAME, L_CRIT, "<cwmp> node has invalid type: %d (expected %d)\n",
					data->type, XML_ELEMENT_NODE);
		return -1;
	}

	// Check name pointer before dereferencing
	if (data->name == NULL)
	{
		log_message(NAME, L_CRIT, "<cwmp> node has NULL name\n");
		return -1;
	}

	// Verify the node name is correct
	if (strcmp((const char *)data->name, "cwmp") != 0)
	{
		log_message(NAME, L_CRIT, "<cwmp> node has wrong name: %s\n", (const char *)data->name);
		return -1;
	}

	log_message(NAME, L_DEBUG, "Successfully validated <cwmp> node (name: %s, type: %d)\n",
				(const char *)data->name, data->type);

	// Log children of <cwmp> for debugging
	if (data->children)
	{
		for (xmlNodePtr child = data->children; child; child = child->next)
		{
			if (child->name)
			{
				log_message(NAME, L_DEBUG, "Child of <cwmp>: node %p (name: %s, type: %d)\n",
							(void*)child, (const char *)child->name, child->type);
			}
			else
			{
				log_message(NAME, L_DEBUG, "Child of <cwmp>: node %p (name: NULL, type: %d)\n",
							(void*)child, child->type);
			}
		}
	}
	else
	{
		log_message(NAME, L_DEBUG, "<cwmp> has no children\n");
	}

	// Find first download element
	b = xmlFindElementByName(data, "download");
	if (!b)
	{
		log_message(NAME, L_NOTICE, "No <download> element found in <cwmp>\n");
		return 0; // No download elements is not an error
	}

	// Process all download elements
	while (b)
	{
		xmlNodePtr next_download = NULL;

		// Validate the download node
		if (!b || b->type != XML_ELEMENT_NODE || !b->name ||
			strcmp((const char *)b->name, "download") != 0)
		{
			log_message(NAME, L_CRIT, "Invalid <download> node encountered\n");
			goto error;
		}

		log_message(NAME, L_DEBUG, "Processing <download> node at %p\n", (void*)b);

		// Initialize all pointers to NULL for this iteration
		command_key = NULL;
		username = NULL;
		password = NULL;
		file_type = NULL;

		// Extract command_key
		c = xmlFindElementByName(b, "command_key");
		if (!c)
		{
			log_message(NAME, L_CRIT, "No <command_key> in <download>\n");
			goto error;
		}
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content)
		{
			xmlNodePtr child = c->children;
			val = xml_get_value_with_whitespace(&child, c);
			command_key = val;
		}
		else
		{
			command_key = strdup("");
		}
		if (!command_key)
		{
			log_message(NAME, L_CRIT, "Failed to allocate memory for command_key\n");
			goto error;
		}

		// Extract URL
		c = xmlFindElementByName(b, "url");
		if (!c)
		{
			log_message(NAME, L_CRIT, "No <url> in <download>\n");
			goto error;
		}
		if (c->children && c->children->content)
		{
			download_url = (char *)c->children->content;
		}
		else
		{
			download_url = "";
		}

		// Extract username
		c = xmlFindElementByName(b, "username");
		if (!c)
		{
			log_message(NAME, L_CRIT, "No <username> in <download>\n");
			goto error;
		}
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content)
		{
			xmlNodePtr child = c->children;
			val = xml_get_value_with_whitespace(&child, c);
			username = val;
		}
		else
		{
			username = strdup("");
		}
		if (!username)
		{
			log_message(NAME, L_CRIT, "Failed to allocate memory for username\n");
			goto error;
		}

		// Extract password
		c = xmlFindElementByName(b, "password");
		if (!c)
		{
			log_message(NAME, L_CRIT, "No <password> in <download>\n");
			goto error;
		}
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content)
		{
			xmlNodePtr child = c->children;
			val = xml_get_value_with_whitespace(&child, c);
			password = val;
		}
		else
		{
			password = strdup("");
		}
		if (!password)
		{
			log_message(NAME, L_CRIT, "Failed to allocate memory for password\n");
			goto error;
		}

		// Extract file_size
		c = xmlFindElementByName(b, "file_size");
		if (!c)
		{
			log_message(NAME, L_CRIT, "No <file_size> in <download>\n");
			goto error;
		}
		if (c->children && c->children->content)
		{
			file_size = (char *)c->children->content;
		}
		else
		{
			file_size = "";
		}

		// Extract time_execute and calculate delay
		c = xmlFindElementByName(b, "time_execute");
		if (!c)
		{
			log_message(NAME, L_CRIT, "No <time_execute> in <download>\n");
			goto error;
		}
		delay = 0; // Default delay
		if (c->children && c->children->content)
		{
			if (sscanf((const char *)c->children->content, "%u", &t) == 1)
			{
				delay = t - time(NULL);
			}
			else
			{
				log_message(NAME, L_WARNING, "Invalid time_execute value: %s\n",
							(const char *)c->children->content);
			}
		}

		// Extract file_type
		c = xmlFindElementByName(b, "file_type");
		if (!c)
		{
			log_message(NAME, L_CRIT, "No <file_type> in <download>\n");
			goto error;
		}
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content)
		{
			xmlNodePtr child = c->children;
			val = xml_get_value_with_whitespace(&child, c);
			file_type = val;
		}
		else
		{
			file_type = strdup("");
		}
		if (!file_type)
		{
			log_message(NAME, L_CRIT, "Failed to allocate memory for file_type\n");
			goto error;
		}

		// Add the download to cwmp
		log_message(NAME, L_DEBUG, "Adding download: key=%s, delay=%d, size=%s, url=%s, type=%s\n",
					command_key, delay, file_size, download_url, file_type);

		cwmp_add_download(command_key, delay, file_size, download_url, file_type, username, password, b);

		// Clean up allocated memory for this iteration
		if (command_key)
		{
			free(command_key);
			command_key = NULL;
		}
		if (username)
		{
			free(username);
			username = NULL;
		}
		if (password)
		{
			free(password);
			password = NULL;
		}
		if (file_type)
		{
			free(file_type);
			file_type = NULL;
		}

		// Find next download element safely
		for (xmlNodePtr tmp = xmlWalkNextOne(b); tmp; tmp = xmlWalkNextOne(tmp))
		{
			if (tmp->type == XML_ELEMENT_NODE && tmp->name &&
				strcmp((const char *)tmp->name, "download") == 0)
			{
				next_download = tmp;
				break;
			}
		}
		b = next_download;
	}

	log_message(NAME, L_DEBUG, "Successfully loaded all downloads\n");
	return 0;

error:
	// Clean up any allocated memory
	if (command_key)
		free(command_key);
	if (username)
		free(username);
	if (password)
		free(password);
	if (file_type)
		free(file_type);

	log_message(NAME, L_CRIT, "Error loading downloads\n");
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
	if (!data)
		return -1;

	// Find first upload element
	b = xmlFindElementByName(data, "upload");

	while (b)
	{
		xmlNodePtr next_upload = NULL;

		c = xmlFindElementByName(b, "command_key");
		if (!c)
			return -1;
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content)
		{
			{
				xmlNodePtr child = c->children;
				val = xml_get_value_with_whitespace(&child, c);
			}
			command_key = val;
		}
		else
			command_key = strdup("");

		c = xmlFindElementByName(b, "url");
		if (!c)
			goto error;
		if (c->children && c->children->content)
			upload_url = (char *)c->children->content;
		else
			upload_url = "";

		c = xmlFindElementByName(b, "username");
		if (!c)
			goto error;
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content)
		{
			{
				xmlNodePtr child = c->children;
				val = xml_get_value_with_whitespace(&child, c);
			}
			username = val;
		}
		else
			username = strdup("");

		c = xmlFindElementByName(b, "password");
		if (!c)
			goto error;
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content)
		{
			{
				xmlNodePtr child = c->children;
				val = xml_get_value_with_whitespace(&child, c);
			}
			password = val;
		}
		else
			password = strdup("");

		c = xmlFindElementByName(b, "time_execute");
		if (!c)
			goto error;
		if (c->children && c->children->content)
		{
			sscanf((const char *)c->children->content, "%u", &t);
			delay = t - time(NULL);
		}

		c = xmlFindElementByName(b, "file_type");
		if (!c)
			goto error;
		if (c->children && c->children->type == XML_TEXT_NODE && c->children->content)
		{
			{
				xmlNodePtr child = c->children;
				val = xml_get_value_with_whitespace(&child, c);
			}
			file_type = val;
		}
		else
			file_type = strdup("");

		cwmp_add_upload(command_key, delay, upload_url, file_type, username, password, b);
		free(command_key);
		free(username);
		free(password);
		free(file_type);

		// Find next upload element
		for (xmlNodePtr tmp = xmlWalkNextOne(b); tmp; tmp = xmlWalkNextOne(tmp))
		{
			if (tmp->type == XML_ELEMENT_NODE && !strcmp((const char *)tmp->name, "upload"))
			{
				next_upload = tmp;
				break;
			}
		}
		b = next_upload;
	}
	return 0;
error:
	free(command_key);
	free(username);
	free(password);
	free(file_type);
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
	if (!data)
		goto error;
	n = xmlNewChild(data, NULL, (const xmlChar *)"event", NULL);
	if (!n)
		goto error;

	if (asprintf(&e, "%d", code) == -1)
		goto error;
	b = xmlNewChild(n, NULL, (const xmlChar *)"event_number", NULL);
	if (!b)
		goto error;
	if (!xmlNewOpaque(b, e))
		goto error;

	if (key)
	{
		b = xmlNewChild(n, NULL, (const xmlChar *)"event_key", NULL);
		if (!b)
			goto error;
		if (!xmlNewOpaque(b, key))
			goto error;
	}

	if (method_id)
	{
		if (asprintf(&c, "%d", method_id) == -1)
			goto error;
		b = xmlNewChild(n, NULL, (const xmlChar *)"event_method_id", NULL);
		if (!b)
			goto error;
		if (!xmlNewOpaque(b, c))
			goto error;
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
	if (!data)
		return -1;

	// Find first event element
	b = xmlFindElementByName(data, "event");

	while (b)
	{
		xmlNodePtr next_event = NULL;

		c = xmlFindElementByName(b, "event_number");
		if (!c || !c->children || c->children->type != XML_TEXT_NODE)
			return -1;
		event_num = (char *)c->children->content;

		c = xmlFindElementByName(b, "event_key");
		if (c && c->children && c->children->type == XML_TEXT_NODE && c->children->content)
		{
			{
				xmlNodePtr child = c->children;
				key = xml_get_value_with_whitespace(&child, c);
			}
		}
		else
			key = NULL;

		c = xmlFindElementByName(b, "event_method_id");
		if (c && c->children && c->children->type == XML_TEXT_NODE && c->children->content)
			method_id = atoi((const char *)c->children->content);

		if (event_num)
		{
			if (e = cwmp_add_event(atoi(event_num), key, method_id, EVENT_NO_BACKUP))
				e->backup_node = b;
			cwmp_add_inform_timer();
		}
		free(key);

		// Find next event element
		for (xmlNodePtr tmp = xmlWalkNextOne(b); tmp; tmp = xmlWalkNextOne(tmp))
		{
			if (tmp->type == XML_ELEMENT_NODE && !strcmp((const char *)tmp->name, "event"))
			{
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
	if (backup_doc)
	{
		xmlFreeDoc(backup_doc);
		backup_doc = NULL;
		backup_tree = NULL;
	}
	xmlCleanupParser();
}