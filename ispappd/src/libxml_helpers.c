#include <string.h>
#include <stdlib.h>
#include "libxml_helpers.h"

// // Helper function to find an element by name (replacement for mxmlFindElement)
// xmlNodePtr xmlFindElementByName(xmlNodePtr node, const char *name)
// {
//     xmlNodePtr curr;
    
//     if (!node || !name)
//         return NULL;
        
//     // Check if this is the node we're looking for
//     if (node->name && !strcmp((const char*)node->name, name))
//         return node;
        
//     // Check children
//     for (curr = node->children; curr; curr = curr->next) {
//         xmlNodePtr found = xmlFindElementByName(curr, name);
//         if (found)
//             return found;
//     }
    
//     return NULL;
// }

// Helper function for node traversal (replacement for mxmlWalkNext)
xmlNodePtr xmlWalkNextOne(xmlNodePtr node)
{
    if (!node) 
        return NULL;
        
    // First try children
    if (node->children)
        return node->children;
        
    // Then try siblings
    if (node->next)
        return node->next;
        
    // Try parent's siblings
    for (xmlNodePtr p = node->parent; p; p = p->parent) {
        if (p->next)
            return p->next;
    }
    
    return NULL;
}

// Helper function to load XML from string (replacement for mxmlLoadString)
xmlDocPtr xmlLoadStringDoc(const char *buffer)
{
    if (!buffer)
        return NULL;
    
    return xmlParseDoc((const xmlChar*)buffer);
}

// Helper function to save XML to string (replacement for mxmlSaveAllocString)
char* xmlSaveString(xmlNodePtr node)
{
    xmlBufferPtr buf;
    char* result;
    
    if (!node)
        return NULL;
    
    buf = xmlBufferCreate();
    if (!buf)
        return NULL;
    
    xmlNodeDump(buf, node->doc, node, 0, 1);
    result = strdup((char*)buf->content);
    xmlBufferFree(buf);
    
    return result;
}

// Helper function to create a text node and add it to parent
// (replacement for mxmlNewOpaque)
xmlNodePtr xmlNewOpaque(xmlNodePtr parent, const char *content)
{
    xmlNodePtr text_node;
    
    if (!parent || !content)
        return NULL;
    
    text_node = xmlNewText((const xmlChar*)content);
    if (!text_node)
        return NULL;
    
    xmlAddChild(parent, text_node);
    return text_node;
}

xmlNodePtr xmlWalkNext(xmlNodePtr node, xmlNodePtr top, int descend)
{
	if (node == NULL)
		return NULL;

	if (node->children != NULL && descend)
		return node->children;

	if (node->next != NULL)
		return node->next;

	for (xmlNodePtr parent = node->parent; parent != NULL; parent = parent->parent)
	{
		if (parent == top)
			return NULL;
		if (parent->next != NULL)
			return parent->next;
	}

	return NULL;
}