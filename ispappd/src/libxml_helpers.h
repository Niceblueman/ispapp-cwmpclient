#ifndef _LIBXML_HELPERS_H
#define _LIBXML_HELPERS_H

#include <libxml2/libxml/tree.h>
#include <libxml2/libxml/parser.h>

// Helper functions to migrate from microxml to libxml2
xmlNodePtr xmlFindElementByName(xmlNodePtr node, const char *name);
xmlNodePtr xmlWalkNext(xmlNodePtr node);
xmlDocPtr xmlLoadStringDoc(const char *buffer);
char* xmlSaveString(xmlNodePtr node);
xmlNodePtr xmlNewOpaque(xmlNodePtr parent, const char *content);

#endif // _LIBXML_HELPERS_H
