#ifndef _LIBXML_HELPERS_H
#define _LIBXML_HELPERS_H

#include <libxml/tree.h>
#include <libxml/parser.h>

// Helper functions to migrate from microxml to libxml2
// xmlNodePtr xmlFindElementByName(xmlNodePtr node, const char *name);
xmlNodePtr xmlWalkNextOne(xmlNodePtr node);
xmlNodePtr xmlWalkNext(xmlNodePtr node, xmlNodePtr top, int descend);
xmlDocPtr xmlLoadStringDoc(const char *buffer);
char* xmlSaveString(xmlNodePtr node);
xmlNodePtr xmlNewOpaque(xmlNodePtr parent, const char *content);

#endif // _LIBXML_HELPERS_H
