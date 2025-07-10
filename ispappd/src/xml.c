#include <stdbool.h>
#include <libubox/uloop.h>
#include <libxml/tree.h>

#include "xml.h"
#include "backup.h"
#include "config.h"
#include "cwmp.h"
#include "external.h"
#include "ispappcwmp.h"
#include "messages.h"
#include "time.h"
#include "json.h"
#include "log.h"

struct fault_code fault_array[]=
{
	[FAULT_0]	 = {"0", "", ""},
	[FAULT_9000] = {"9000", "Server", "Method not supported"},
	[FAULT_9001] = {"9001", "Server", "Request denied"},
	[FAULT_9002] = {"9002", "Server", "Internal error"},
	[FAULT_9003] = {"9003", "Client", "Invalid arguments"},
	[FAULT_9004] = {"9004", "Server", "Resources exceeded"},
	[FAULT_9005] = {"9005", "Client", "Invalid parameter name"},
	[FAULT_9006] = {"9006", "Client", "Invalid parameter type"},
	[FAULT_9007] = {"9007", "Client", "Invalid parameter value"},
	[FAULT_9008] = {"9008", "Client", "Attempt to set a non-writable parameter"},
	[FAULT_9009] = {"9009", "Server", "Notification request rejected"},
	[FAULT_9010] = {"9010", "Server", "Download failure"},
	[FAULT_9011] = {"9011", "Server", "Upload failure"},
	[FAULT_9012] = {"9012", "Server", "File transfer server authentication failure"},
	[FAULT_9013] = {"9013", "Server", "Unsupported protocol for file transfer"},
	[FAULT_9014] = {"9014", "Server", "Download failure: unable to join multicast group"},
	[FAULT_9015] = {"9015", "Server", "Download failure: unable to contact file server"},
	[FAULT_9016] = {"9016", "Server", "Download failure: unable to access file"},
	[FAULT_9017] = {"9017", "Server", "Download failure: unable to complete download"},
	[FAULT_9018] = {"9018", "Server", "Download failure: file corrupted"},
	[FAULT_9019] = {"9019", "Server", "Download failure: file authentication failure"}
};

const static char *soap_env_url = "http://schemas.xmlsoap.org/soap/envelope/";
const static char *soap_enc_url = "http://schemas.xmlsoap.org/soap/encoding/";
const static char *xsd_url = "http://www.w3.org/2001/XMLSchema";
const static char *xsi_url = "http://www.w3.org/2001/XMLSchema-instance";
const static char *cwmp_urls[] = {
		"urn:dslforum-org:cwmp-1-0", 
		"urn:dslforum-org:cwmp-1-1", 
		"urn:dslforum-org:cwmp-1-2", 
		NULL };

static struct cwmp_namespaces ns;

const struct rpc_method rpc_methods[] = {
	{ "GetRPCMethods", xml_handle_get_rpc_methods },
	{ "SetParameterValues", xml_handle_set_parameter_values },
	{ "GetParameterValues", xml_handle_get_parameter_values },
	{ "GetParameterNames", xml_handle_get_parameter_names },
	{ "GetParameterAttributes", xml_handle_get_parameter_attributes },
	{ "SetParameterAttributes", xml_handle_set_parameter_attributes },
	{ "AddObject", xml_handle_AddObject },
	{ "DeleteObject", xml_handle_DeleteObject },
	{ "Download", xml_handle_download },
	{ "Upload", xml_handle_upload },
	{ "Reboot", xml_handle_reboot },
	{ "FactoryReset", xml_handle_factory_reset },
	{ "ScheduleInform", xml_handle_schedule_inform },
};

xmlNodePtr				/* O - Element node or NULL */
xmlFindNodeWithContent(xmlNodePtr node,	/* I - Current node */
						xmlNodePtr top,	/* I - Top node */
						const char *text,	/* I - Element text, if NULL return NULL */
						int descend)		/* I - Descend into tree */
{
	if (!node || !top || !text)
		return (NULL);

	// Navigate through the tree
	xmlNodePtr current = node;
	
	while (current != NULL) {
		// Check if current node has the text content we're looking for
		if (current->type == XML_TEXT_NODE && 
			current->content &&
			(!text || !strcmp((char*)current->content, text)))
		{
			return (current);
		}

		// Process child nodes if descend is requested
		if (descend && current->children != NULL) {
			xmlNodePtr result = xmlFindNodeWithContent(current->children, top, text, descend);
			if (result != NULL)
				return result;
		}

		// Move to next sibling
		current = current->next;
	}
	return (NULL);
}

const char *xml_format_cb(xmlNodePtr node, int pos)
{
	xmlNodePtr b = node;
	static char space_format[20];
	int i=0;

	// In libxml2, we don't have the same whitespace positions as in microxml
	// We'll define our own constants for similar positions
	#define XML_WS_BEFORE_CLOSE 1
	#define XML_WS_BEFORE_OPEN 2
	#define XML_WS_AFTER_OPEN 3
	#define XML_WS_AFTER_CLOSE 4

	switch (pos) {
		case XML_WS_BEFORE_CLOSE:
			if (node->children && node->children->type != XML_ELEMENT_NODE)
				return ("");
			
			while (b->parent != NULL) {
				space_format[i] = ' ';
				b = b->parent;
				i++;
			}
			space_format[i] = '\0';
			return (space_format);
				
		case XML_WS_BEFORE_OPEN:
			while (b->parent != NULL) {
				space_format[i] = ' ';
				b = b->parent;
				i++;
			}
			space_format[i] = '\0';
			return (space_format);
		case XML_WS_AFTER_OPEN:
			if (node->children && node->children->type != XML_ELEMENT_NODE)
				return ("");
			else
				return ("\n");
		case XML_WS_AFTER_CLOSE:
			return ("\n");
		default:
			return ("");
	}
}

char *xml_get_value_with_whitespace(xmlNodePtr *b, xmlNodePtr body_in)
{
	char *value = NULL;
	// In libxml2, node content is directly available as node->content
	if (*b && (*b)->content)
		value = strdup((char *)((*b)->content));
	return value;
}

static inline void xml_free_ns(void)
{
	int i = 0;
	FREE(ns.soap_enc);
	FREE(ns.xsd);
	FREE(ns.xsi);
	FREE(ns.cwmp);
	for (i = 0; i < ARRAY_SIZE(ns.soap_env) && ns.soap_env[i]; i++) {
		FREE(ns.soap_env[i]);
	}
}

void xml_exit(void)
{
	xml_free_ns();
}

void xml_log_parameter_fault()
{
	struct list_head *ilist;
	struct external_parameter *external_parameter;

	list_for_each_prev(ilist, &external_list_parameter) {
		external_parameter = list_entry(ilist, struct external_parameter, list);
		if (external_parameter->fault_code && external_parameter->fault_code[0]=='9') {
			log_message(NAME, L_NOTICE, "Fault in the param: %s , Fault code: %s\n", external_parameter->name, external_parameter->fault_code);
		}
		else {
			break;
		}
	}
}

int xml_check_duplicated_parameter(xmlNodePtr tree)
{
	xmlNodePtr currentNode, compareNode;
	
	// Helper function to traverse the XML tree
	xmlNodePtr xmlWalkNext(xmlNodePtr node, xmlNodePtr top, int descend) {
		if (node == NULL)
			return NULL;
			
		if (node->children != NULL && descend)
			return node->children;
			
		if (node->next != NULL)
			return node->next;
			
		for (xmlNodePtr parent = node->parent; parent != NULL; parent = parent->parent) {
			if (parent == top)
				return NULL;
			if (parent->next != NULL)
				return parent->next;
		}
		
		return NULL;
	}
	
	// Start traversal
	currentNode = tree;
	while (currentNode) {
		// Check if current node is a text node under a "Name" element
		if (currentNode && currentNode->type == XML_TEXT_NODE && 
			currentNode->content &&
			currentNode->parent &&
			currentNode->parent->type == XML_ELEMENT_NODE &&
			!xmlStrcmp(currentNode->parent->name, (const xmlChar*)"Name")) {
			
			// Compare with other Name nodes
			compareNode = currentNode;
			while ((compareNode = xmlWalkNext(compareNode, tree, 1))) {
				if (compareNode && compareNode->type == XML_TEXT_NODE &&
					compareNode->content &&
					compareNode->parent && 
					compareNode->parent->type == XML_ELEMENT_NODE &&
					!xmlStrcmp(compareNode->parent->name, (const xmlChar*)"Name")) {
						
					if (!xmlStrcmp(compareNode->content, currentNode->content)) {
						log_message(NAME, L_NOTICE, "Fault in the param: %s, Fault code: 9003 <parameter duplicated>\n", compareNode->content);
						return 1;
					}
				}
			}
		}
		currentNode = xmlWalkNext(currentNode, tree, 1);
	}
	return 0;
}

// Helper function to find elements by name
xmlNodePtr xmlFindElementByName(xmlNodePtr node, const char *name)
{
    if (!node || !name)
        return NULL;
        
    // Check if the current node is what we're looking for
    if (node->type == XML_ELEMENT_NODE && !xmlStrcmp(node->name, (const xmlChar*)name))
        return node;
        
    // Check children
    xmlNodePtr child = node->children;
    while (child) {
        xmlNodePtr result = xmlFindElementByName(child, name);
        if (result)
            return result;
        child = child->next;
    }
    
    return NULL;
}

// Find element by name with attribute
xmlNodePtr xmlFindElementWithAttr(xmlNodePtr node, const char *name, const char *attr_name, const char *attr_value)
{
    if (!node || !name || !attr_name || !attr_value)
        return NULL;
        
    // Check current node
    if (node->type == XML_ELEMENT_NODE && 
        !xmlStrcmp(node->name, (const xmlChar*)name)) {
        
        xmlChar *attr = xmlGetProp(node, (const xmlChar*)attr_name);
        if (attr && !xmlStrcmp(attr, (const xmlChar*)attr_value)) {
            xmlFree(attr);
            return node;
        }
        if (attr) xmlFree(attr);
    }
    
    // Check children
    xmlNodePtr child = node->children;
    while (child) {
        xmlNodePtr result = xmlFindElementWithAttr(child, name, attr_name, attr_value);
        if (result)
            return result;
        child = child->next;
    }
    
    return NULL;
}

int xml_get_attrname_array(xmlNodePtr node,
                             const char *value,
                             char *name_arr[],
                             int size)
{
    int j = 0;
    xmlAttrPtr attr;

    if (!node || node->type != XML_ELEMENT_NODE || !value)
        return (-1);

    for (attr = node->properties; attr != NULL; attr = attr->next)
    {
        xmlChar *attr_value = xmlGetProp(node, attr->name);
        if (attr_value) {
            if (!xmlStrcmp(attr_value, (const xmlChar*)value) && 
                xmlStrlen(attr->name) > 5 && 
                *(attr->name + 5) == ':')
            {
                name_arr[j++] = xmlStrdup(attr->name + 6);
            }
            xmlFree(attr_value);
            if (j >= size) break;
        }
    }

    return (j ? 0 : -1);
}

xmlNodePtr xml_find_node_by_env_type(xmlNodePtr tree_in, char *bname) {
    xmlNodePtr b;
    char *c;
    int i;

    for (i = 0; i < ARRAY_SIZE(ns.soap_env) && ns.soap_env[i]; i++) {
        if (asprintf(&c, "%s:%s", ns.soap_env[i], bname) == -1)
            return NULL;

        // Recursive search in the tree for this element
        b = NULL;
        for (xmlNodePtr curr = tree_in; curr != NULL; curr = xmlWalkNext(curr, tree_in, 1)) {
            if (curr->type == XML_ELEMENT_NODE && !xmlStrcmp(curr->name, (const xmlChar*)c)) {
                FREE(c);
                return curr;
            }
        }
        FREE(c);
    }
    return NULL;
}

// Helper function for walking the XML tree - similar to mxmlWalkNext
xmlNodePtr xmlWalkNext(xmlNodePtr node, xmlNodePtr top, int descend)
{
    if (!node)
        return NULL;
        
    if (descend && node->children)
        return node->children;
        
    if (node->next)
        return node->next;
        
    for (xmlNodePtr parent = node->parent; parent != NULL; parent = parent->parent) {
        if (parent == top)
            return NULL;
        if (parent->next)
            return parent->next;
    }
    
    return NULL;
}

// Helper function to find attribute by value and get the attribute name
char* xmlGetAttrNameByValue(xmlNodePtr node, const char* value)
{
    xmlAttrPtr attr;
    for (attr = node->properties; attr; attr = attr->next) {
        xmlChar* attrVal = xmlGetProp(node, attr->name);
        if (attrVal && !xmlStrcmp(attrVal, (const xmlChar*)value)) {
            char* result = NULL;
            if (xmlStrlen(attr->name) > 5 && (*(attr->name + 5) == ':')) {
                result = strdup((const char*)(attr->name + 6));
            }
            xmlFree(attrVal);
            return result;
        }
        if (attrVal) xmlFree(attrVal);
    }
    return NULL;
}

static int xml_recreate_namespace(xmlNodePtr tree)
{
    xmlNodePtr b = tree;
    const char *cwmp_urn;
    char *c;
    int i;

    xml_free_ns();

    do {
        if (ns.cwmp == NULL) {
            for (i = 0; cwmp_urls[i] != NULL; i++) {
                cwmp_urn = cwmp_urls[i];
                c = xmlGetAttrNameByValue(b, cwmp_urn);
                if (c) {
                    ns.cwmp = c; // c is already strdup'd in xmlGetAttrNameByValue
                    break;
                }
            }
        }

        if (ns.soap_env[0] == NULL) {
            xml_get_attrname_array(b, soap_env_url, ns.soap_env, ARRAY_SIZE(ns.soap_env));
        }

        if (ns.soap_enc == NULL) {
            c = xmlGetAttrNameByValue(b, soap_enc_url);
            if (c) {
                ns.soap_enc = c; // c is already strdup'd in xmlGetAttrNameByValue
            }
        }

        if (ns.xsd == NULL) {
            c = xmlGetAttrNameByValue(b, xsd_url);
            if (c) {
                ns.xsd = c; // c is already strdup'd in xmlGetAttrNameByValue
            }
        }

        if (ns.xsi == NULL) {
            c = xmlGetAttrNameByValue(b, xsi_url);
            if (c) {
                ns.xsi = c; // c is already strdup'd in xmlGetAttrNameByValue
            }
        }
    } while ((b = xmlWalkNext(b, tree, 1)) != NULL);

    if ((ns.soap_env[0] != NULL) && (ns.cwmp != NULL))
        return 0;

    return -1;
}

static void xml_get_hold_request(xmlNodePtr tree)
{
    xmlNodePtr b;
    char *c;

    cwmp->hold_requests = false;

    if (asprintf(&c, "%s:%s", ns.cwmp, "NoMoreRequests") == -1)
        return;
    
    // Find the element
    b = NULL;
    for (xmlNodePtr curr = tree; curr != NULL; curr = xmlWalkNext(curr, tree, 1)) {
        if (curr->type == XML_ELEMENT_NODE && !xmlStrcmp(curr->name, (const xmlChar*)c)) {
            b = curr;
            break;
        }
    }
    free(c);
    
    if (b) {
        // Get the text content
        xmlNodePtr text_node = b->children;
        if (text_node && text_node->type == XML_TEXT_NODE && text_node->content) {
            cwmp->hold_requests = (atoi((const char*)text_node->content)) ? true : false;
        }
    }

    if (asprintf(&c, "%s:%s", ns.cwmp, "HoldRequests") == -1)
        return;
    
    // Find the element
    b = NULL;
    for (xmlNodePtr curr = tree; curr != NULL; curr = xmlWalkNext(curr, tree, 1)) {
        if (curr->type == XML_ELEMENT_NODE && !xmlStrcmp(curr->name, (const xmlChar*)c)) {
            b = curr;
            break;
        }
    }
    free(c);
    
    if (b) {
        // Get the text content
        xmlNodePtr text_node = b->children;
        if (text_node && text_node->type == XML_TEXT_NODE && text_node->content) {
            cwmp->hold_requests = (atoi((const char*)text_node->content)) ? true : false;
        }
    }
}

int xml_handle_message(char *msg_in, char **msg_out)
{
    xmlDocPtr doc_in = NULL, doc_out = NULL;
    xmlNodePtr tree_in = NULL, tree_out = NULL, b, body_out;
    const struct rpc_method *method;
    int i, code = FAULT_9002;
    char *c;

    // Parse the response template
    doc_out = xmlParseMemory(CWMP_RESPONSE_MESSAGE, strlen(CWMP_RESPONSE_MESSAGE));
    if (!doc_out) goto error;
    tree_out = xmlDocGetRootElement(doc_out);
    if (!tree_out) goto error;

    // Parse the incoming message
    doc_in = xmlParseMemory(msg_in, strlen(msg_in));
    if (!doc_in) goto error;
    tree_in = xmlDocGetRootElement(doc_in);
    if (!tree_in) goto error;

    if (xml_recreate_namespace(tree_in)) {
        code = FAULT_9003;
        goto fault_out;
    }
    
    /* handle cwmp:ID */
    if (asprintf(&c, "%s:%s", ns.cwmp, "ID") == -1)
        goto error;

    // Find ID element
    b = NULL;
    for (xmlNodePtr curr = tree_in; curr != NULL; curr = xmlWalkNext(curr, tree_in, 1)) {
        if (curr->type == XML_ELEMENT_NODE && !xmlStrcmp(curr->name, (const xmlChar*)c)) {
            b = curr;
            break;
        }
    }
    FREE(c);
    
    /* ACS did not send ID parameter, we are continuing without it */
    if (!b) goto find_method;

    // Get the text content
    xmlNodePtr text_node = b->children;
    if (!text_node || text_node->type != XML_TEXT_NODE || !text_node->content) goto find_method;
    c = strdup((char*)text_node->content);

    // Find ID element in output document
    b = NULL;
    for (xmlNodePtr curr = tree_out; curr != NULL; curr = xmlWalkNext(curr, tree_out, 1)) {
        if (curr->type == XML_ELEMENT_NODE && !xmlStrcmp(curr->name, (const xmlChar*)"cwmp:ID")) {
            b = curr;
            break;
        }
    }
    if (!b) {
        FREE(c);
        goto error;
    }

    // Set ID in output document
    xmlNodePtr new_text = xmlNewText((const xmlChar*)c);
    if (!new_text) {
        FREE(c);
        goto error;
    }
    xmlAddChild(b, new_text);
    FREE(c);

find_method:
    b = xml_find_node_by_env_type(tree_in, "Body");
    if (!b) {
        code = FAULT_9003;
        goto fault_out;
    }
    
    // Find the first element node child of Body
    xmlNodePtr child = b->children;
    while (child) {
        if (child->type == XML_ELEMENT_NODE) {
            b = child;
            break;
        }
        child = child->next;
    }
    
    if (!child) {
        code = FAULT_9003;
        goto fault_out;
    }

    // Get node name
    c = (char*)b->name;
    if (strchr(c, ':')) {
        char *tmp = strchr(c, ':');
        size_t ns_len = tmp - c;

        if (strlen(ns.cwmp) != ns_len) {
            code = FAULT_9003;
            goto fault_out;
        }

        if (strncmp(ns.cwmp, c, ns_len)) {
            code = FAULT_9003;
            goto fault_out;
        }

        c = tmp + 1;
    } else {
        code = FAULT_9003;
        goto fault_out;
    }
    
    method = NULL;
    log_message(NAME, L_NOTICE, "received %s method from the ACS\n", c);
    for (i = 0; i < ARRAY_SIZE(rpc_methods); i++) {
        if (!strcmp(c, rpc_methods[i].name)) {
            method = &rpc_methods[i];
            break;
        }
    }
    
    if (method) {
        if (method->handler(b, tree_in, tree_out)) goto error;
    }
    else {
        code = FAULT_9000;
        goto fault_out;
    }
    
    // Serialize the output document
    xmlBufferPtr buf = xmlBufferCreate();
    if (!buf) goto error;
    
    xmlNodeDumpOutput(xmlOutputBufferCreateBuffer(buf, NULL), doc_out, tree_out, 0, 1, NULL);
    *msg_out = strdup((char*)xmlBufferContent(buf));
    xmlBufferFree(buf);

    xmlFreeDoc(doc_in);
    xmlFreeDoc(doc_out);
    return 0;

fault_out:
    body_out = NULL;
    for (xmlNodePtr curr = tree_out; curr != NULL; curr = xmlWalkNext(curr, tree_out, 1)) {
        if (curr->type == XML_ELEMENT_NODE && !xmlStrcmp(curr->name, (const xmlChar*)"soap_env:Body")) {
            body_out = curr;
            break;
        }
    }
    if (!body_out) goto error;
    
    xml_create_generic_fault_message(body_out, code);
    
    // Serialize the output document
    xmlBufferPtr buf = xmlBufferCreate();
    if (!buf) goto error;
    
    xmlNodeDumpOutput(xmlOutputBufferCreateBuffer(buf, NULL), doc_out, tree_out, 0, 1, NULL);
    *msg_out = strdup((char*)xmlBufferContent(buf));
    xmlBufferFree(buf);

    xmlFreeDoc(doc_in);
    xmlFreeDoc(doc_out);
    return 0;

error:
    if (doc_in) xmlFreeDoc(doc_in);
    if (doc_out) xmlFreeDoc(doc_out);
    return -1;
}

int xml_get_index_fault(char *fault_code)
{
	int i;

	for (i = 0; i < __FAULT_MAX; i++) {
		if (strcmp(fault_array[i].code, fault_code) == 0)
			return i;
	}
	return FAULT_9002;
}

int xml_check_fault_in_list_parameter(void)
{
	struct external_parameter *external_parameter;
	struct list_head *ilist;
	int code;

	ilist = external_list_parameter.prev;
	if (ilist != &external_list_parameter) {
		external_parameter = list_entry(ilist, struct external_parameter, list);
		if (external_parameter->fault_code && external_parameter->fault_code[0] == '9') {
			code = xml_get_index_fault(external_parameter->fault_code);
			return code;
		}
	}
	return 0;
}

/* Inform */

static int xml_prepare_events_inform(xmlNodePtr tree)
{
    xmlNodePtr node, b1 = NULL, b2;
    char *c;
    int n = 0;
    struct list_head *p;
    struct event *event;

    // Find Event element in the tree
    for (xmlNodePtr curr = tree; curr != NULL; curr = xmlWalkNext(curr, tree, 1)) {
        if (curr->type == XML_ELEMENT_NODE && !xmlStrcmp(curr->name, (const xmlChar*)"Event")) {
            b1 = curr;
            break;
        }
    }
    if (!b1) return -1;

    list_for_each(p, &cwmp->events) {
        event = list_entry(p, struct event, list);
        
        // Create EventStruct element
        node = xmlNewChild(b1, NULL, (const xmlChar*)"EventStruct", NULL);
        if (!node) goto error;

        // Create EventCode element
        b2 = xmlNewChild(node, NULL, (const xmlChar*)"EventCode", NULL);
        if (!b2) goto error;
        
        // Add the event code text
        xmlNodePtr text = xmlNewText((const xmlChar*)event_code_array[event->code].code);
        if (!text) goto error;
        xmlAddChild(b2, text);

        // Create CommandKey element
        b2 = xmlNewChild(node, NULL, (const xmlChar*)"CommandKey", NULL);
        if (!b2) goto error;

        // Add CommandKey text if it exists
        if (event->key) {
            text = xmlNewText((const xmlChar*)event->key);
            if (!text) goto error;
            xmlAddChild(b2, text);
        }

        n++;
    }

    if (n) {
        if (asprintf(&c, "cwmp:EventStruct[%u]", n) == -1)
            return -1;

        xmlSetProp(b1, (const xmlChar*)"soap_enc:arrayType", (const xmlChar*)c);
        FREE(c);
    }

    return 0;

error:
    return -1;
}

static int xml_prepare_notifications_inform(xmlNodePtr parameter_list, int *counter)
{
    /* notifications */
    xmlNodePtr b, n;
    xmlNodePtr text;

    struct list_head *p;
    struct notification *notification;

    list_for_each(p, &cwmp->notifications) {
        notification = list_entry(p, struct notification, list);

        // Check if parameter already exists in list
        b = xmlFindNodeWithContent(parameter_list, parameter_list, notification->parameter, 1);
        if (b) continue;
        
        // Create new ParameterValueStruct
        n = xmlNewChild(parameter_list, NULL, (const xmlChar*)"ParameterValueStruct", NULL);
        if (!n) goto error;

        // Create Name element
        b = xmlNewChild(n, NULL, (const xmlChar*)"Name", NULL);
        if (!b) goto error;

        // Add parameter name as text
        text = xmlNewText((const xmlChar*)notification->parameter);
        if (!text) goto error;
        xmlAddChild(b, text);

        // Create Value element
        b = xmlNewChild(n, NULL, (const xmlChar*)"Value", NULL);
        if (!b) goto error;

        // Set xsi:type attribute
        xmlSetProp(b, (const xmlChar*)"xsi:type", (const xmlChar*)notification->type);

        // Add parameter value as text
        text = xmlNewText((const xmlChar*)notification->value);
        if (!text) goto error;
        xmlAddChild(b, text);

        (*counter)++;
    }

    return 0;

error:
    return -1;
}

int xml_prepare_inform_message(char **msg_out)
{
    xmlDocPtr doc = NULL;
    xmlNodePtr tree, b, n, parameter_list;
    struct external_parameter *external_parameter;
    xmlNodePtr text_node;
    char *c;
    int counter = 0;

    // Parse the template
    doc = xmlParseMemory(CWMP_INFORM_MESSAGE, strlen(CWMP_INFORM_MESSAGE));
    if (!doc) goto error;
    tree = xmlDocGetRootElement(doc);
    if (!tree) goto error;

    if (xml_add_cwmpid(tree)) goto error;

    // Find RetryCount element
    b = xmlFindElementByName(tree, "RetryCount");
    if (!b) goto error;

    // Set RetryCount value
    char retry_count_str[16];
    snprintf(retry_count_str, sizeof(retry_count_str), "%d", cwmp->retry_count);
    text_node = xmlNewText((const xmlChar*)retry_count_str);
    if (!text_node) goto error;
    xmlAddChild(b, text_node);

    // Find Manufacturer element
    b = xmlFindElementByName(tree, "Manufacturer");
    if (!b) goto error;

    // Set Manufacturer value
    text_node = xmlNewText((const xmlChar*)cwmp->deviceid.manufacturer);
    if (!text_node) goto error;
    xmlAddChild(b, text_node);

    // Find OUI element
    b = xmlFindElementByName(tree, "OUI");
    if (!b) goto error;

    // Set OUI value
    text_node = xmlNewText((const xmlChar*)cwmp->deviceid.oui);
    if (!text_node) goto error;
    xmlAddChild(b, text_node);

    // Find ProductClass element
    b = xmlFindElementByName(tree, "ProductClass");
    if (!b) goto error;

    // Set ProductClass value
    text_node = xmlNewText((const xmlChar*)cwmp->deviceid.product_class);
    if (!text_node) goto error;
    xmlAddChild(b, text_node);

    // Find SerialNumber element
    b = xmlFindElementByName(tree, "SerialNumber");
    if (!b) goto error;

    // Set SerialNumber value
    text_node = xmlNewText((const xmlChar*)cwmp->deviceid.serial_number);
    if (!text_node) goto error;
    xmlAddChild(b, text_node);
   
    if (xml_prepare_events_inform(tree))
        goto error;

    // Find CurrentTime element
    b = xmlFindElementByName(tree, "CurrentTime");
    if (!b) goto error;

    // Set CurrentTime value
    text_node = xmlNewText((const xmlChar*)mix_get_time());
    if (!text_node) goto error;
    xmlAddChild(b, text_node);

    external_action_simple_execute("inform", "parameter", NULL);
    if (external_action_handle(json_handle_get_parameter_value))
        goto error;

    // Find ParameterList element
    parameter_list = xmlFindElementByName(tree, "ParameterList");
    if (!parameter_list) goto error;

    while (external_list_parameter.next != &external_list_parameter) {
        external_parameter = list_entry(external_list_parameter.next, struct external_parameter, list);

        // Create ParameterValueStruct element
        n = xmlNewChild(parameter_list, NULL, (const xmlChar*)"ParameterValueStruct", NULL);
        if (!n) goto error;

        // Create Name element
        b = xmlNewChild(n, NULL, (const xmlChar*)"Name", NULL);
        if (!b) goto error;

        // Set Name value
        text_node = xmlNewText((const xmlChar*)external_parameter->name);
        if (!text_node) goto error;
        xmlAddChild(b, text_node);

        // Create Value element
        b = xmlNewChild(n, NULL, (const xmlChar*)"Value", NULL);
        if (!b) goto error;

        // Set xsi:type attribute
        xmlSetProp(b, (const xmlChar*)"xsi:type", (const xmlChar*)external_parameter->type);
        
        // Set Value content
        text_node = xmlNewText((const xmlChar*)(external_parameter->data ? external_parameter->data : ""));
        if (!text_node) goto error;
        xmlAddChild(b, text_node);

        counter++;

        external_parameter_delete(external_parameter);
    }

    if (xml_prepare_notifications_inform(parameter_list, &counter))
        goto error;

    if (asprintf(&c, "cwmp:ParameterValueStruct[%d]", counter) == -1)
        goto error;

    xmlSetProp(parameter_list, (const xmlChar*)"soap_enc:arrayType", (const xmlChar*)c);
    FREE(c);

    // Serialize XML to string
    xmlBufferPtr buf = xmlBufferCreate();
    if (!buf) goto error;
    
    xmlNodeDumpOutput(xmlOutputBufferCreateBuffer(buf, NULL), doc, tree, 0, 1, NULL);
    *msg_out = strdup((char*)xmlBufferContent(buf));
    xmlBufferFree(buf);

    xmlFreeDoc(doc);
    return 0;

error:
    external_free_list_parameter();
    if (doc) xmlFreeDoc(doc);
    return -1;
}

int xml_parse_inform_response_message(char *msg_in)
{
    xmlDocPtr doc = NULL;
    xmlNodePtr tree, b;
    int fault = 0;

    // Parse the XML
    doc = xmlParseMemory(msg_in, strlen(msg_in));
    if (!doc) goto error;
    tree = xmlDocGetRootElement(doc);
    if (!tree) goto error;
    
    // Process namespace info
    if(xml_recreate_namespace(tree)) goto error;

    // Find Fault element
    b = xml_find_node_by_env_type(tree, "Fault");
    if (b) {
        // Look for fault code 8005
        xmlNodePtr fault_node = NULL;
        for (xmlNodePtr curr = b; curr != NULL; curr = xmlWalkNext(curr, tree, 1)) {
            if (curr->type == XML_TEXT_NODE && 
                curr->content &&
                !strcmp((char*)curr->content, "8005")) {
                fault = FAULT_ACS_8005;
                goto out;
            }
        }
        goto error;
    }

    // Process hold request info
    xml_get_hold_request(tree);
    
    // Find MaxEnvelopes element
    b = xmlFindElementByName(tree, "MaxEnvelopes");
    if (!b) goto error;

    // Get text content
    xmlNodePtr text_node = b->children;
    if (!text_node || text_node->type != XML_TEXT_NODE || !text_node->content)
        goto error;

out:
    xmlFreeDoc(doc);
    return fault;

error:
    if (doc) xmlFreeDoc(doc);
    return -1;
}

/* ACS GetRPCMethods */
int xml_prepare_get_rpc_methods_message(char **msg_out)
{
    xmlDocPtr doc = NULL;
    xmlNodePtr tree;

    // Parse the template
    doc = xmlParseMemory(CWMP_GET_RPC_METHOD_MESSAGE, strlen(CWMP_GET_RPC_METHOD_MESSAGE));
    if (!doc) return -1;
    tree = xmlDocGetRootElement(doc);
    if (!tree) {
        xmlFreeDoc(doc);
        return -1;
    }

    if(xml_add_cwmpid(tree)) {
        xmlFreeDoc(doc);
        return -1;
    }

    // Serialize XML to string
    xmlBufferPtr buf = xmlBufferCreate();
    if (!buf) {
        xmlFreeDoc(doc);
        return -1;
    }
    
    xmlNodeDumpOutput(xmlOutputBufferCreateBuffer(buf, NULL), doc, tree, 0, 1, NULL);
    *msg_out = strdup((char*)xmlBufferContent(buf));
    xmlBufferFree(buf);

    xmlFreeDoc(doc);
    return 0;
}

int xml_parse_get_rpc_methods_response_message(char *msg_in)
{
    xmlDocPtr doc = NULL;
    xmlNodePtr tree, b;
    int fault = 0;

    // Parse XML
    doc = xmlParseMemory(msg_in, strlen(msg_in));
    if (!doc) goto error;
    tree = xmlDocGetRootElement(doc);
    if (!tree) goto error;
    
    if(xml_recreate_namespace(tree)) goto error;

    // Find Fault element
    b = xml_find_node_by_env_type(tree, "Fault");
    if (b) {
        // Check for fault code 8005
        xmlNodePtr fault_node = NULL;
        for (xmlNodePtr curr = b; curr != NULL; curr = xmlWalkNext(curr, tree, 1)) {
            if (curr->type == XML_TEXT_NODE && 
                curr->content &&
                !strcmp((char*)curr->content, "8005")) {
                fault = FAULT_ACS_8005;
                goto out;
            }
        }
        goto out;
    }

    xml_get_hold_request(tree);

out:
    xmlFreeDoc(doc);
    return fault;

error:
    if (doc) xmlFreeDoc(doc);
    return -1;
}

/* ACS TransferComplete */

int xml_parse_transfer_complete_response_message(char *msg_in)
{
    xmlDocPtr doc = NULL;
    xmlNodePtr tree, b;
    int fault = 0;

    // Parse XML
    doc = xmlParseMemory(msg_in, strlen(msg_in));
    if (!doc) goto error;
    tree = xmlDocGetRootElement(doc);
    if (!tree) goto error;
    
    if(xml_recreate_namespace(tree)) goto error;

    // Find Fault element
    b = xml_find_node_by_env_type(tree, "Fault");
    if (b) {
        // Check for fault code 8005
        xmlNodePtr fault_node = NULL;
        for (xmlNodePtr curr = b; curr != NULL; curr = xmlWalkNext(curr, tree, 1)) {
            if (curr->type == XML_TEXT_NODE && 
                curr->content &&
                !strcmp((char*)curr->content, "8005")) {
                fault = FAULT_ACS_8005;
                goto out;
            }
        }
        goto out;
    }

    xml_get_hold_request(tree);

out:
    xmlFreeDoc(doc);
    return fault;

error:
    if (doc) xmlFreeDoc(doc);
    return -1;
}

/* CPE GetRPCMethods */

static int xml_handle_get_rpc_methods(xmlNodePtr body_in,
					xmlNodePtr tree_in,
					xmlNodePtr tree_out)
{
		xmlNodePtr b1, b2, method_list;
		int i = 0;

		// Find the SOAP body element
		// xmlFindElementHelper is a custom helper function that would need to be implemented
		b1 = xmlFindElementByName(tree_out, "soap_env:Body");
		if (!b1) return -1;

		// Create new element nodes
		b1 = xmlNewChild(b1, NULL, BAD_CAST "cwmp:GetRPCMethodsResponse", NULL);
		if (!b1) return -1;

		method_list = xmlNewChild(b1, NULL, BAD_CAST "MethodList", NULL);
		if (!method_list) return -1;

		for (i = 0; i < ARRAY_SIZE(rpc_methods); i++) {
			b2 = xmlNewChild(method_list, NULL, BAD_CAST "string", BAD_CAST rpc_methods[i].name);
			if (!b2) return -1;
		}
        
		// Set attribute on the method_list element
		char *attr_value;
		if (asprintf(&attr_value, "xsd:string[%d]", ARRAY_SIZE(rpc_methods)) == -1)
			return -1;

		xmlNewProp(method_list, BAD_CAST "soap_enc:arrayType", BAD_CAST attr_value);
		free(attr_value);

		log_message(NAME, L_NOTICE, "send GetRPCMethodsResponse to the ACS\n");
		return 0;
}

/* SetParameterValues */

int xml_handle_set_parameter_values(xmlNodePtr body_in,
                            xmlNodePtr tree_in,
                            xmlNodePtr tree_out)
{
    xmlNodePtr b, body_out;
    struct external_parameter *external_parameter;
    struct list_head *ilist;
    char *parameter_name = NULL, *parameter_value = NULL, *status = NULL, *param_key = NULL;
    int code = FAULT_9002;

    // Find the body element in the output tree
    body_out = xmlFindElementByName(tree_out, "soap_env:Body");
    if (!body_out) goto error;

    // Check for duplicated parameters
    if (xml_check_duplicated_parameter(body_in)) {
        code = FAULT_9003;
        goto fault_out;
    }

    // Traverse the XML tree to find parameter names, values, and the parameter key
    for (b = body_in; b != NULL; b = xmlWalkNext(b, tree_in, 1)) {
        // Find parameter name
        if (b && b->type == XML_TEXT_NODE && 
            b->content &&
            b->parent && b->parent->type == XML_ELEMENT_NODE && 
            !xmlStrcmp(b->parent->name, (const xmlChar*)"Name")) {
            parameter_name = (char*)b->content;
        }
        
        // Handle empty Name element
        if (b && b->type == XML_ELEMENT_NODE && 
            !xmlStrcmp(b->name, (const xmlChar*)"Name") && 
            !b->children) {
            parameter_name = "";
        }
        
        // Find parameter value
        if (b && b->type == XML_TEXT_NODE && 
            b->content &&
            b->parent && b->parent->type == XML_ELEMENT_NODE && 
            !xmlStrcmp(b->parent->name, (const xmlChar*)"Value")) {
            free(parameter_value);
            parameter_value = xml_get_value_with_whitespace(&b, body_in);
        }
        
        // Handle empty Value element
        if (b && b->type == XML_ELEMENT_NODE && 
            !xmlStrcmp(b->name, (const xmlChar*)"Value") && 
            !b->children) {
            free(parameter_value);
            parameter_value = strdup("");
        }
        
        // Find parameter key
        if (b && b->type == XML_TEXT_NODE && 
            b->content &&
            b->parent && b->parent->type == XML_ELEMENT_NODE && 
            !xmlStrcmp(b->parent->name, (const xmlChar*)"ParameterKey")) {
            free(param_key);
            param_key = xml_get_value_with_whitespace(&b, body_in);
        }
        
        // Handle empty ParameterKey element
        if (b && b->type == XML_ELEMENT_NODE && 
            !xmlStrcmp(b->name, (const xmlChar*)"ParameterKey") && 
            !b->children) {
            free(param_key);
            param_key = strdup("");
        }

        // Process parameter if both name and value are available
        if (parameter_name && parameter_value) {
            external_action_parameter_execute("set", "value", parameter_name, parameter_value);
            parameter_name = NULL;
            FREE(parameter_value);
        }
    }

    // Apply the settings with the parameter key
    external_action_simple_execute("apply", "value", param_key);
    free(param_key);
    
    // Handle any errors from the external action
    if (external_action_handle(json_handle_set_parameter))
        goto fault_out;

    // Check for faults in the parameter list
    if (xml_check_fault_in_list_parameter()) {
        code = FAULT_9003;
        goto fault_out;
    }
    
    // Fetch the response status
    external_fetch_set_param_resp_status(&status);
    if(!status)
        goto fault_out;

    // Create the response element
    b = xmlNewChild(body_out, NULL, (const xmlChar*)"cwmp:SetParameterValuesResponse", NULL);
    if (!b) goto error;

    // Create the Status element
    xmlNodePtr status_node = xmlNewChild(b, NULL, (const xmlChar*)"Status", NULL);
    if (!status_node) goto error;

    // Set the Status value
    xmlNodePtr text_node = xmlNewText((const xmlChar*)status);
    if (!text_node) goto error;
    xmlAddChild(status_node, text_node);

    free(status);
    free(parameter_value);
    external_free_list_parameter();

    log_message(NAME, L_NOTICE, "send SetParameterValuesResponse to the ACS\n");
    return 0;

fault_out:
    xml_log_parameter_fault();
    free(parameter_value);
    xml_create_set_parameter_value_fault_message(body_out, code);
    free(status);
    external_free_list_parameter();
    return 0;

error:
    free(parameter_value);
    free(status);
    external_free_list_parameter();
    return -1;
}

/* GetParameterValues */

int xml_handle_get_parameter_values(xmlNodePtr body_in,
                             xmlNodePtr tree_in,
                             xmlNodePtr tree_out)
{
    xmlNodePtr n, parameter_list, b, body_out, t;
    struct external_parameter *external_parameter;
    char *parameter_name = NULL;
    int counter = 0, fc, code = FAULT_9002;
    struct list_head *ilist;
    xmlNodePtr text_node;

    // Find body element in output tree
    body_out = xmlFindElementByName(tree_out, "soap_env:Body");
    if (!body_out) return -1;

    // Traverse the XML tree to find parameter names
    for (b = body_in; b != NULL; b = xmlWalkNext(b, tree_in, 1)) {
        // Find parameter name in "string" element
        if (b && b->type == XML_TEXT_NODE && 
            b->content &&
            b->parent && b->parent->type == XML_ELEMENT_NODE && 
            !xmlStrcmp(b->parent->name, (const xmlChar*)"string")) {
            parameter_name = (char*)b->content;
        }

        // Handle empty "string" element
        if (b && b->type == XML_ELEMENT_NODE && 
            !xmlStrcmp(b->name, (const xmlChar*)"string") && 
            !b->children) {
            parameter_name = "";
        }

        // Process parameter if name is available
        if (parameter_name) {
            external_action_parameter_execute("get", "value", parameter_name, NULL);
            if (external_action_handle(json_handle_get_parameter_value))
                goto fault_out;
            fc = xml_check_fault_in_list_parameter();
            if (fc) {
                code = fc;
                goto fault_out;
            }
        }
        parameter_name = NULL;
    }

    // Create response elements
    n = xmlNewChild(body_out, NULL, (const xmlChar*)"cwmp:GetParameterValuesResponse", NULL);
    if (!n) goto out;
    
    parameter_list = xmlNewChild(n, NULL, (const xmlChar*)"ParameterList", NULL);
    if (!parameter_list) goto out;

    // Process all parameters in the list
    while (external_list_parameter.next != &external_list_parameter) {
        external_parameter = list_entry(external_list_parameter.next, struct external_parameter, list);

        // Create ParameterValueStruct element
        n = xmlNewChild(parameter_list, NULL, (const xmlChar*)"ParameterValueStruct", NULL);
        if (!n) goto out;

        // Create Name element
        t = xmlNewChild(n, NULL, (const xmlChar*)"Name", NULL);
        if (!t) goto out;

        // Set Name value
        text_node = xmlNewText((const xmlChar*)external_parameter->name);
        if (!text_node) goto out;
        xmlAddChild(t, text_node);

        // Create Value element
        t = xmlNewChild(n, NULL, (const xmlChar*)"Value", NULL);
        if (!t) goto out;

        // Set Value type attribute
        xmlSetProp(t, (const xmlChar*)"xsi:type", (const xmlChar*)external_parameter->type);
        
        // Set Value content
        text_node = xmlNewText((const xmlChar*)(external_parameter->data ? external_parameter->data : ""));
        if (!text_node) goto out;
        xmlAddChild(t, text_node);

        counter++;
        external_parameter_delete(external_parameter);
    }
    
    // Set array type attribute
    char *c;
    if (asprintf(&c, "cwmp:ParameterValueStruct[%d]", counter) == -1)
        goto out;

    xmlSetProp(parameter_list, (const xmlChar*)"soap_enc:arrayType", (const xmlChar*)c);
    FREE(c);

    log_message(NAME, L_NOTICE, "send GetParameterValuesResponse to the ACS\n");
    return 0;
    
fault_out:
    xml_log_parameter_fault();
    xml_create_generic_fault_message(body_out, code);
    external_free_list_parameter();
    return 0;
    
out:
    external_free_list_parameter();
    return -1;
}

/* GetParameterNames */

int xml_handle_get_parameter_names(xmlNodePtr body_in,
					xmlNodePtr tree_in,
					xmlNodePtr tree_out)
{
	xmlNodePtr n, parameter_list, b = body_in, body_out, t;
	struct external_parameter *external_parameter;
	char *parameter_name = NULL;
	char *next_level = NULL;
	int counter = 0, fc, code = FAULT_9002;

	body_out = xmlFindElementByName(tree_out, "soap_env:Body");
	if (!body_out) return -1;
	
	while (b) {
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "ParameterPath")) {
			parameter_name = (char*)b->content;
		}

		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "ParameterPath") &&
			!b->children) {
			parameter_name = "";
		}

		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "NextLevel")) {
			next_level = (char*)b->content;
		}

		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "NextLevel") &&
			!b->children) {
			next_level = "";
		}

		b = xmlWalkNext(b);
	}
	
	if (parameter_name && next_level) {
		external_action_parameter_execute("get", "name", parameter_name, next_level);
		if (external_action_handle(json_handle_get_parameter_name))
			goto fault_out;
		fc = xml_check_fault_in_list_parameter();
		if (fc) {
			code = fc;
			goto fault_out;
		}
	}

	n = xmlNewChild(body_out, NULL, (const xmlChar*)"cwmp:GetParameterNamesResponse", NULL);
	if (!n) goto out;

	parameter_list = xmlNewChild(n, NULL, (const xmlChar*)"ParameterList", NULL);
	if (!parameter_list) goto out;

	while (external_list_parameter.next != &external_list_parameter) {
		external_parameter = list_entry(external_list_parameter.next, struct external_parameter, list);

		n = xmlNewChild(parameter_list, NULL, (const xmlChar*)"ParameterInfoStruct", NULL);
		if (!n) goto out;

		t = xmlNewChild(n, NULL, (const xmlChar*)"Name", NULL);
		if (!t) goto out;

		xmlNodePtr text_node = xmlNewText((const xmlChar*)external_parameter->name);
		if (!text_node) goto out;
		xmlAddChild(t, text_node);

		t = xmlNewChild(n, NULL, (const xmlChar*)"Writable", NULL);
		if (!t) goto out;

		text_node = xmlNewText((const xmlChar*)external_parameter->data);
		if (!text_node) goto out;
		xmlAddChild(t, text_node);

		counter++;

		external_parameter_delete(external_parameter);
	}

	char *c;
	if (asprintf(&c, "cwmp:ParameterInfoStruct[%d]", counter) == -1)
		goto out;

	xmlSetProp(parameter_list, (const xmlChar*)"soap_enc:arrayType", (const xmlChar*)c);
	FREE(c);

	log_message(NAME, L_NOTICE, "send GetParameterNamesResponse to the ACS\n");
	return 0;
fault_out:
	xml_log_parameter_fault();
	xml_create_generic_fault_message(body_out, code);
	external_free_list_parameter();
	return 0;

out:
	external_free_list_parameter();
	return -1;
}

/* GetParameterAttributes */

static int xml_handle_get_parameter_attributes(xmlNodePtr body_in,
					xmlNodePtr tree_in,
					xmlNodePtr tree_out)
{
	xmlNodePtr n, parameter_list, b = body_in, body_out, t;
	struct external_parameter *external_parameter;
	char *parameter_name = NULL;
	int counter = 0, fc, code = FAULT_9002;
	struct list_head *ilist;

	body_out = xmlFindElementByName(tree_out, "soap_env:Body");
	if (!body_out) return -1;

	while (b) {
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "string")) {
			parameter_name = (char*)b->content;
		}

		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "string") &&
			!b->children) {
			parameter_name = "";
		}
		if (parameter_name) {
			external_action_parameter_execute("get", "notification", parameter_name, NULL);
			if (external_action_handle(json_handle_get_parameter_attribute))
				goto fault_out;
			fc = xml_check_fault_in_list_parameter();
			if (fc) {
				code = fc;
				goto fault_out;
			}
		}
		b = xmlWalkNext(b);
		parameter_name = NULL;
	}

	n = xmlNewChild(body_out, NULL, (const xmlChar*)"cwmp:GetParameterAttributesResponse", NULL);
	if (!n) goto out;

	parameter_list = xmlNewChild(n, NULL, (const xmlChar*)"ParameterList", NULL);
	if (!parameter_list) goto out;

	while (external_list_parameter.next != &external_list_parameter) {

		external_parameter = list_entry(external_list_parameter.next, struct external_parameter, list);

		n = xmlNewChild(parameter_list, NULL, (const xmlChar*)"ParameterAttributeStruct", NULL);
			if (!n) goto out;

		t = xmlNewChild(n, NULL, (const xmlChar*)"Name", NULL);
		if (!t) goto out;

		xmlNodePtr text_node = xmlNewText((const xmlChar*)external_parameter->name);
		if (!text_node) goto out;
		xmlAddChild(t, text_node);

		t = xmlNewChild(n, NULL, (const xmlChar*)"Notification", NULL);
		if (!t) goto out;
		
		text_node = xmlNewText((const xmlChar*)(external_parameter->data ? external_parameter->data : ""));
		if (!text_node) goto out;
		xmlAddChild(t, text_node);

		t = xmlNewChild(n, NULL, (const xmlChar*)"AccessList", NULL);
		if (!t) goto out;

		counter++;

		external_parameter_delete(external_parameter);
	}
	char *c;
	if (asprintf(&c, "cwmp:ParameterAttributeStruct[%d]", counter) == -1)
		goto out;

	xmlSetProp(parameter_list, (const xmlChar*)"soap_enc:arrayType", (const xmlChar*)c);
	FREE(c);

	log_message(NAME, L_NOTICE, "send GetParameterAttributesResponse to the ACS\n");
	return 0;
fault_out:
	xml_log_parameter_fault();
	xml_create_generic_fault_message(body_out, code);
	external_free_list_parameter();
	return 0;
out:
	external_free_list_parameter();
	return -1;
}

/* SetParameterAttributes */

static int xml_handle_set_parameter_attributes(xmlNodePtr body_in,
						xmlNodePtr tree_in,
						xmlNodePtr tree_out) {

	xmlNodePtr b = body_in, body_out;
	char *c, *parameter_name = NULL, *parameter_notification = NULL, *success = NULL;
	uint8_t attr_notification_update = 0;
	struct external_parameter *external_parameter;
	struct list_head *ilist;
	int fc, code = FAULT_9002 ;

	body_out = xmlFindElementByName(tree_out, "soap_env:Body");
	if (!body_out) goto error;

	while (b != NULL) {
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "SetParameterAttributesStruct")) {
			attr_notification_update = 0;
			parameter_name = NULL;
			parameter_notification = NULL;
		}
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "Name")) {
			parameter_name = (char*)b->content;
		}

		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "Name") &&
			!b->children) {
			parameter_name = "";
		}

		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "NotificationChange")) {
			if (strcasecmp((const char*)b->content, "true") == 0) {
				attr_notification_update = 1;
			} else if (strcasecmp((const char*)b->content, "false") == 0) {
				attr_notification_update = 0;
			} else {
				attr_notification_update = (uint8_t) atoi((const char*)b->content);
			}
		}

		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "Notification")) {
			parameter_notification = (char*)b->content;
		}

		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "Notification") &&
			!b->children) {
			parameter_notification = "";
		}

		if (attr_notification_update && parameter_name && parameter_notification) {
			external_action_parameter_execute("set", "notification", parameter_name, parameter_notification);
			attr_notification_update = 0;
			parameter_name = NULL;
			parameter_notification = NULL;
		}
		b = xmlWalkNext(b);
	}

	external_action_simple_execute("apply", "notification", NULL);

	if (external_action_handle(json_handle_set_parameter))
		goto fault_out;

	fc = xml_check_fault_in_list_parameter();
	if (fc) {
		code = fc;
		goto fault_out;
	}

	external_fetch_set_param_resp_status(&success);
	if(!success)
		goto fault_out;

	b = xmlNewChild(body_out, NULL, (const xmlChar*)"cwmp:SetParameterAttributesResponse", NULL);
	if (!b) goto error;

	free(success);
	external_free_list_parameter();

	log_message(NAME, L_NOTICE, "send SetParameterAttributesResponse to the ACS\n");
	return 0;

fault_out:
	xml_log_parameter_fault();
	xml_create_generic_fault_message(body_out, code);
	free(success);
	external_free_list_parameter();
	return 0;
error:
	free(success);
	external_free_list_parameter();
	return -1;
}

/* Download */

static int xml_handle_download(xmlNodePtr body_in,
					xmlNodePtr tree_in,
					xmlNodePtr tree_out)
{
	xmlNodePtr n, t, b = body_in, body_out;
	char *download_url = NULL, *file_size = NULL,
		*command_key = NULL, *file_type = NULL, *username = NULL,
		*password = NULL, r;
	int delay = -1, code = FAULT_9002;

	body_out = xmlFindElementByName(tree_out, "soap_env:Body");
	if (!body_out) return -1;

	while (b != NULL) {
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "CommandKey")) {
			FREE(command_key);
			command_key = xml_get_value_with_whitespace(&b, body_in);
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "CommandKey") &&
			!b->children) {
			FREE(command_key);
			command_key = strdup("");
		}
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "FileType")) {
			FREE(file_type);
			file_type = xml_get_value_with_whitespace(&b, body_in);
		}
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "URL")) {
			download_url = (char*)b->content;
		}
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "Username")) {
			FREE(username);
			username = xml_get_value_with_whitespace(&b, body_in);
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "Username") &&
			!b->children) {
			FREE(username);
			username = strdup("");
		}
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "Password")) {
			FREE(password);
			password = xml_get_value_with_whitespace(&b, body_in);
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "Password") &&
			!b->children) {
			FREE(password);
			password = strdup("");
		}
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "FileSize")) {
			file_size = (char*)b->content;
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "FileSize") &&
			!b->children) {
			file_size = "0";
		}
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "DelaySeconds")) {
			delay = atoi((const char*)b->content);
		}
		b = xmlWalkNext(b);
	}
	if (!download_url || !file_size || !command_key || !file_type || !username || !password || delay < 0) {
		code = FAULT_9003;
		goto fault_out;
	}
	if (sscanf(download_url,"%*[a-zA-Z_0-9]://%c",&r) < 1 ||
		sscanf(download_url,"%*[^:]://%*[^:]:%*[^@]@%c",&r) == 1) {
		code = FAULT_9003;
		goto fault_out;
	}
	if (cwmp->download_count >= MAX_DOWNLOAD) {
		code = FAULT_9004;
		goto fault_out;
	}
	n = backup_add_download(command_key, delay, file_size, download_url, file_type, username, password);
	cwmp_add_download(command_key, delay, file_size, download_url, file_type, username, password, n);
	FREE(file_type);
	FREE(command_key);
	FREE(username);
	FREE(password);

	t = xmlNewChild(body_out, NULL, (const xmlChar*)"cwmp:DownloadResponse", NULL);
	if (!t) return -1;

	b = xmlNewChild(t, NULL, (const xmlChar*)"Status", NULL);
	if (!b) return -1;

	b = xmlNewChild(t, NULL, (const xmlChar*)"StartTime", NULL);
	if (!b) return -1;

	xmlNodePtr text_node = xmlNewText((const xmlChar*)UNKNOWN_TIME);
	if (!text_node) return -1;
	xmlAddChild(b, text_node);

	b = xmlFindElementByName(t, "Status");
	if (!b) return -1;

	text_node = xmlNewText((const xmlChar*)"1");
	if (!text_node) return -1;
	xmlAddChild(b, text_node);

	b = xmlNewChild(t, NULL, (const xmlChar*)"CompleteTime", NULL);
	if (!b) return -1;

	text_node = xmlNewText((const xmlChar*)UNKNOWN_TIME);
	if (!text_node) return -1;
	xmlAddChild(b, text_node);
	if (!b) return -1;

	log_message(NAME, L_NOTICE, "send DownloadResponse to the ACS\n");
	return 0;

fault_out:
	xml_create_generic_fault_message(body_out, code);
	FREE(file_type);
	FREE(command_key);
	FREE(username);
	FREE(password);
	return 0;
}


/* upload */

static int xml_handle_upload(xmlNodePtr body_in,
					xmlNodePtr tree_in,
					xmlNodePtr tree_out)
{
	xmlNodePtr n, t, b = body_in, body_out;
	char *upload_url = NULL,
		*command_key = NULL, *file_type = NULL, *username = NULL,
		*password = NULL, r;
	int delay = -1, code = FAULT_9002;

	body_out = xmlFindElementByName(tree_out, "soap_env:Body");
	if (!body_out) {
		printf("!body_out) \n" );
		return -1;
	}

	while (b != NULL) {
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "CommandKey")) {
			FREE(command_key);
			command_key = xml_get_value_with_whitespace(&b, body_in);
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "CommandKey") &&
			!b->children) {
			FREE(command_key);
			command_key = strdup("");
		}
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "FileType")) {
			FREE(file_type);
			file_type = xml_get_value_with_whitespace(&b, body_in);
		}
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "URL")) {
			upload_url = (char*)b->content;
		}
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "Username")) {
			FREE(username);
			username = xml_get_value_with_whitespace(&b, body_in);
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "Username") &&
			!b->children) {
			FREE(username);
			username = strdup("");
		}
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "Password")) {
			FREE(password);
			password = xml_get_value_with_whitespace(&b, body_in);
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "Password") &&
			!b->children) {
			FREE(password);
			password = strdup("");
		}
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "DelaySeconds")) {
			delay = atoi((const char*)b->content);
		}
		b = xmlWalkNext(b);
	}
	if (!upload_url || !command_key || !file_type || !username || !password || delay < 0) {
		code = FAULT_9003;
		goto fault_out;
	}
	if (sscanf(upload_url,"%*[a-zA-Z_0-9]://%c",&r) < 1 ||
		sscanf(upload_url,"%*[^:]://%*[^:]:%*[^@]@%c",&r) == 1) {
		code = FAULT_9003;
		goto fault_out;
	}
	if (cwmp->upload_count >= MAX_UPLOAD) {
		code = FAULT_9004;
		goto fault_out;
	}
	n = backup_add_upload(command_key, delay, upload_url, file_type, username, password);
	cwmp_add_upload(command_key, delay, upload_url, file_type, username, password, n);
	FREE(file_type);
	FREE(command_key);
	FREE(username);
	FREE(password);

	t = xmlNewChild(body_out, NULL, (const xmlChar*)"cwmp:UploadResponse", NULL);
	if (!t) return -1;

	b = xmlNewChild(t, NULL, (const xmlChar*)"Status", NULL);
	if (!b) return -1;

	b = xmlNewChild(t, NULL, (const xmlChar*)"StartTime", NULL);
	if (!b) return -1;

	xmlNodePtr text_node = xmlNewText((const xmlChar*)UNKNOWN_TIME);
	if (!text_node) return -1;
	xmlAddChild(b, text_node);

	b = xmlFindElementByName(t, "Status");
	if (!b) return -1;

	text_node = xmlNewText((const xmlChar*)"1");
	if (!text_node) return -1;
	xmlAddChild(b, text_node);

	b = xmlNewChild(t, NULL, (const xmlChar*)"CompleteTime", NULL);
	if (!b) return -1;

	text_node = xmlNewText((const xmlChar*)UNKNOWN_TIME);
	if (!text_node) return -1;
	xmlAddChild(b, text_node);
	if (!b) return -1;

	return 0;

fault_out:
	xml_create_generic_fault_message(body_out, code);
	FREE(file_type);
	FREE(command_key);
	FREE(username);
	FREE(password);
	return 0;
}


/* FactoryReset */

static int xml_handle_factory_reset(xmlNodePtr node,
					xmlNodePtr tree_in,
					xmlNodePtr tree_out)
{
	xmlNodePtr body_out, b;

	body_out = xmlFindElementByName(tree_out, "soap_env:Body");
	if (!body_out) return -1;

	b = xmlNewChild(body_out, NULL, (const xmlChar*)"cwmp:FactoryResetResponse", NULL);
	if (!b) return -1;

	cwmp_add_handler_end_session(ENDS_FACTORY_RESET);

	log_message(NAME, L_NOTICE, "send FactoryResetResponse to the ACS\n");
	return 0;
}

 /* Reboot */

static int xml_handle_reboot(xmlNodePtr node,
					xmlNodePtr tree_in,
					xmlNodePtr tree_out)
{
	xmlNodePtr b = node, body_out;
	char *command_key = NULL;
	int code = FAULT_9002;

	body_out = xmlFindElementByName(tree_out, "soap_env:Body");
	if (!body_out) return -1;

	while (b) {
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "CommandKey")) {
			FREE(command_key);
			command_key = xml_get_value_with_whitespace(&b, node);
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "CommandKey") &&
			!b->children) {
			FREE(command_key);
			command_key = strdup("");
		}
		b = xmlWalkNext(b);
	}

	if (!command_key) {
		code = FAULT_9003;
		goto fault_out;
	}

	b = xmlNewChild(body_out, NULL, (const xmlChar*)"cwmp:RebootResponse", NULL);
	if (!b) {
		FREE(command_key);
		return -1;
	}

	backup_add_event(EVENT_M_REBOOT, command_key, 0);
	cwmp_add_handler_end_session(ENDS_REBOOT);

	FREE(command_key);

	log_message(NAME, L_NOTICE, "send RebootResponse to the ACS\n");
	return 0;

fault_out:
	xml_create_generic_fault_message(body_out, code);
	FREE(command_key);
	return 0;
}

/* ScheduleInform */

static int xml_handle_schedule_inform(xmlNodePtr body_in,
					xmlNodePtr tree_in,
					xmlNodePtr tree_out)
{
	xmlNodePtr b = body_in, body_out;
	char *command_key = NULL;
	char *delay_seconds = NULL;
	int  delay = 0, code = FAULT_9002;

	body_out = xmlFindElementByName(tree_out, "soap_env:Body");
	if (!body_out) return -1;

	while (b) {
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "CommandKey")) {
			FREE(command_key);
			command_key = xml_get_value_with_whitespace(&b, body_in);
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "CommandKey") &&
			!b->children) {
			FREE(command_key);
			command_key = strdup("");
		}

		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "DelaySeconds")) {
			delay_seconds = (char*)b->content;
		}
		b = xmlWalkNext(b);
	}
	if (delay_seconds) delay = atoi(delay_seconds);

	if (command_key && (delay > 0)) {
		cwmp_add_scheduled_inform(command_key, delay);
		b = xmlNewChild(body_out, NULL, (const xmlChar*)"cwmp:ScheduleInformResponse", NULL);
		if (!b) goto error;
	}
	else {
		code = FAULT_9003
		goto fault_out;
	}
	FREE(command_key);
	log_message(NAME, L_NOTICE, "send ScheduleInformResponse to the ACS\n");
	return 0;

fault_out:
	FREE(command_key);
	xml_create_generic_fault_message(body_out, code);
	return 0;

error:
	FREE(command_key);
	return -1;
}

/* AddObject */

static int xml_handle_AddObject(xmlNodePtr body_in,
					xmlNodePtr tree_in,
					xmlNodePtr tree_out)
{
	xmlNodePtr b = body_in, t, body_out;
	char *object_name = NULL, *param_key = NULL;
	char *status = NULL, *fault = NULL, *instance = NULL;
	int code = FAULT_9002;

	body_out = xmlFindElementByName(tree_out, "soap_env:Body");
	if (!body_out) return -1;

	while (b) {
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "ObjectName")) {
			object_name = (char*)b->content;
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "ObjectName") &&
			!b->children) {
			object_name = "";
		}

		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "ParameterKey")) {
			free(param_key);
			param_key = xml_get_value_with_whitespace(&b, body_in);
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "ParameterKey") &&
			!b->children) {
			free(param_key);
			param_key = strdup("");
		}
		b = xmlWalkNext(b);
	}

	if (!param_key) {
		code = FAULT_9003;
		goto fault_out;
	}

	if (object_name) {
		external_action_parameter_execute("add", "object", object_name, NULL);
		if (external_action_handle(json_handle_add_object)) goto fault_out;
	} else {
		code = FAULT_9003;
		goto fault_out;
	}

	external_fetch_add_obj_resp(&status, &instance, &fault);

	if (fault && fault[0] == '9') {
		code = xml_get_index_fault(fault);
		goto fault_out;
	}
	if (!status || !instance) {
		code = FAULT_9002;
		goto fault_out;
	}

	external_action_simple_execute("apply", "object", param_key);
	FREE(param_key);

	t = xmlNewChild(body_out, NULL, (const xmlChar*)"cwmp:AddObjectResponse", NULL);
	if (!t) goto error;

	b = xmlNewChild(t, NULL, (const xmlChar*)"InstanceNumber", NULL);
	if (!b) goto error;
	xmlNodePtr text_node = xmlNewText((const xmlChar*)instance);
	if (!text_node) goto error;
	xmlAddChild(b, text_node);

	b = xmlNewChild(t, NULL, (const xmlChar*)"Status", NULL);
	if (!b) goto error;
	text_node = xmlNewText((const xmlChar*)status);
	if (!text_node) goto error;
	xmlAddChild(b, text_node);

	free(instance);
	free(status);
	free(fault);

	log_message(NAME, L_NOTICE, "send AddObjectResponse to the ACS\n");
	return 0;

fault_out:
	log_message(NAME, L_NOTICE, "Fault in the param: %s, Fault code: %s\n", object_name ? object_name : "", fault_array[code].code);
	xml_create_generic_fault_message(body_out, code);
	FREE(param_key);
	free(instance);
	free(status);
	free(fault);
	return 0;

error:
	FREE(param_key);
	free(instance);
	free(status);
	free(fault);
	return -1;
}

/* DeleteObject */

static int xml_handle_DeleteObject(xmlNodePtr body_in,
					xmlNodePtr tree_in,
					xmlNodePtr tree_out)
{
	xmlNodePtr b = body_in, t, body_out;
	char *object_name = NULL, *param_key = NULL;
	char *status = NULL, *fault = NULL;
	int code = FAULT_9002;

	body_out = xmlFindElementByName(tree_out, "soap_env:Body");
	if (!body_out) return -1;

	while (b) {
		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "ObjectName")) {
			object_name = (char*)b->content;
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "ObjectName") &&
			!b->children) {
			object_name = "";
		}

		if (b && b->type == XML_TEXT_NODE &&
			b->content &&
			b->parent->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->parent->name, "ParameterKey")) {
			free(param_key);
			param_key = xml_get_value_with_whitespace(&b, body_in);
		}
		if (b && b->type == XML_ELEMENT_NODE &&
			!strcmp((const char*)b->name, "ParameterKey") &&
			!b->children) {
			free(param_key);
			param_key = strdup("");
		}
		b = xmlWalkNext(b);
	}

	if (!param_key) {
		code = FAULT_9003;
		goto fault_out;
	}

	if (object_name) {
		external_action_parameter_execute("delete", "object", object_name, NULL);
		if (external_action_handle(json_handle_method_status)) goto fault_out;
	} else {
		code = FAULT_9003;
		goto fault_out;
	}

	external_fetch_method_resp_status(&status, &fault);

	if (fault && fault[0] == '9') {
		code = xml_get_index_fault(fault);
		goto fault_out;
	}
	if (!status ) {
		code = FAULT_9002;
		goto fault_out;
	}

	external_action_simple_execute("apply", "object", param_key);
	FREE(param_key);

	t = xmlNewChild(body_out, NULL, (const xmlChar*)"cwmp:DeleteObjectResponse", NULL);
	if (!t) goto error;

	b = xmlNewChild(t, NULL, (const xmlChar*)"Status", NULL);
	if (!b) goto error;
	
	xmlNodePtr text_node = xmlNewText((const xmlChar*)status);
	if (!text_node) goto error;
	xmlAddChild(b, text_node);
	
	free(status);
	free(fault);

	log_message(NAME, L_NOTICE, "send DeleteObjectResponse to the ACS\n");
	return 0;

fault_out:
	log_message(NAME, L_NOTICE, "Fault in the param: %s, Fault code: %s\n", object_name ? object_name : "", fault_array[code].code);
	xml_create_generic_fault_message(body_out, code);
	FREE(param_key);
	free(status);
	free(fault);
	return 0;

error:
	FREE(param_key);
	free(status);
	free(fault);
	return -1;
}

/* Fault */

xmlNodePtr xml_create_generic_fault_message(xmlNodePtr body, int code)
{
    xmlNodePtr b, t;

    // Create Fault element
    b = xmlNewChild(body, NULL, (const xmlChar*)"soap_env:Fault", NULL);
    if (!b) return NULL;

    // Create faultcode element and set content
    t = xmlNewChild(b, NULL, (const xmlChar*)"faultcode", NULL);
    if (!t) return NULL;
    
    xmlNodePtr text_node = xmlNewText((const xmlChar*)fault_array[code].type);
    if (!text_node) return NULL;
    xmlAddChild(t, text_node);

    // Create faultstring element and set content
    t = xmlNewChild(b, NULL, (const xmlChar*)"faultstring", NULL);
    if (!t) return NULL;
    
    text_node = xmlNewText((const xmlChar*)"CWMP fault");
    if (!text_node) return NULL;
    xmlAddChild(t, text_node);

    // Create detail element
    t = xmlNewChild(b, NULL, (const xmlChar*)"detail", NULL);
    if (!t) return NULL;

    // Create cwmp:Fault element
    b = xmlNewChild(t, NULL, (const xmlChar*)"cwmp:Fault", NULL);
    if (!b) return NULL;

    // Create FaultCode element and set content
    t = xmlNewChild(b, NULL, (const xmlChar*)"FaultCode", NULL);
    if (!t) return NULL;
    
    text_node = xmlNewText((const xmlChar*)fault_array[code].code);
    if (!text_node) return NULL;
    xmlAddChild(t, text_node);

    // Create FaultString element and set content
    t = xmlNewChild(b, NULL, (const xmlChar*)"FaultString", NULL);
    if (!t) return NULL;
    
    text_node = xmlNewText((const xmlChar*)fault_array[code].string);
    if (!text_node) return NULL;
    xmlAddChild(t, text_node);

    log_message(NAME, L_NOTICE, "send Fault: %s: '%s'\n", fault_array[code].code, fault_array[code].string);
    return b;
}

int xml_create_set_parameter_value_fault_message(xmlNodePtr body, int code)
{
    struct external_parameter *external_parameter;
    xmlNodePtr b, n, t;
    int index;
    xmlNodePtr text_node;

    n = xml_create_generic_fault_message(body, code);
    if (!n)
        return -1;

    while (external_list_parameter.next != &external_list_parameter) {
        external_parameter = list_entry(external_list_parameter.next, struct external_parameter, list);

        if (external_parameter->fault_code && external_parameter->fault_code[0]=='9') {
            index = xml_get_index_fault(external_parameter->fault_code);

            // Create SetParameterValuesFault element
            b = xmlNewChild(n, NULL, (const xmlChar*)"SetParameterValuesFault", NULL);
            if (!b) return -1;

            // Create ParameterName element
            t = xmlNewChild(b, NULL, (const xmlChar*)"ParameterName", NULL);
            if (!t) return -1;
            
            // Set ParameterName value
            text_node = xmlNewText((const xmlChar*)external_parameter->name);
            if (!text_node) return -1;
            xmlAddChild(t, text_node);

            // Create FaultCode element
            t = xmlNewChild(b, NULL, (const xmlChar*)"FaultCode", NULL);
            if (!t) return -1;
            
            // Set FaultCode value
            text_node = xmlNewText((const xmlChar*)external_parameter->fault_code);
            if (!text_node) return -1;
            xmlAddChild(t, text_node);

            // Create FaultString element
            t = xmlNewChild(b, NULL, (const xmlChar*)"FaultString", NULL);
            if (!t) return -1;
            
            // Set FaultString value
            text_node = xmlNewText((const xmlChar*)fault_array[index].string);
            if (!text_node) return -1;
            xmlAddChild(t, text_node);
        }
        external_parameter_delete(external_parameter);
    }
    return 0;
}

int xml_add_cwmpid(xmlNodePtr tree)
{
    xmlNodePtr b;
    static unsigned int id = 0;
    char buf[16];
    
    // Find the cwmp:ID element
    b = NULL;
    for (xmlNodePtr curr = tree; curr != NULL; curr = xmlWalkNext(curr, tree, 1)) {
        if (curr->type == XML_ELEMENT_NODE && !xmlStrcmp(curr->name, (const xmlChar*)"cwmp:ID")) {
            b = curr;
            break;
        }
    }
    if (!b) return -1;
    
    // Set the ID
    snprintf(buf, sizeof(buf), "%u", ++id);
    xmlNodePtr text = xmlNewText((const xmlChar*)buf);
    if (!text) return -1;
    xmlAddChild(b, text);
    
    return 0;
}