#ifndef XML_COMPAT_H
#define XML_COMPAT_H

#ifdef HAVE_LIBROXML
#include <roxml.h>

// Compatibility types
typedef node_t xml_node_t;

// Compatibility constants
#define MXML_OPAQUE 0
#define MXML_DESCEND 1
#define MXML_NO_DESCEND 0
#define MXML_DESCEND_FIRST 2
#define MXML_WS_BEFORE_OPEN 0
#define MXML_WS_AFTER_OPEN 1
#define MXML_WS_BEFORE_CLOSE 2
#define MXML_WS_AFTER_CLOSE 3
#define MXML_ELEMENT 1
#define MXML_OPAQUE_CALLBACK NULL
#define MXML_NO_CALLBACK NULL

// Compatibility structures
typedef struct {
    char *name;
    char *value;
} mxml_attr_t;

typedef struct {
    char *name;
    int num_attrs;
    mxml_attr_t *attrs;
} mxml_element_t;

typedef struct {
    char *opaque;
} mxml_opaque_t;

typedef union {
    mxml_element_t element;
    mxml_opaque_t opaque;
} mxml_value_t;

// Helper functions for node access
static inline int xml_get_node_type(xml_node_t *node) {
    if (!node) return -1;
    int type = roxml_get_type(node);
    if (type == ROXML_TXT_NODE) return MXML_OPAQUE;
    if (type == ROXML_ELM_NODE) return MXML_ELEMENT;
    return type;
}

static inline const char *xml_get_node_opaque(xml_node_t *node) {
    if (!node) return NULL;
    if (roxml_get_type(node) == ROXML_TXT_NODE) {
        return roxml_get_content(node, NULL, 0, NULL);
    }
    return NULL;
}

static inline const char *xml_get_node_element_name(xml_node_t *node) {
    if (!node) return NULL;
    if (roxml_get_type(node) == ROXML_ELM_NODE) {
        return roxml_get_name(node, NULL, 0);
    }
    return NULL;
}

static inline xml_node_t *xml_get_node_parent(xml_node_t *node) {
    if (!node) return NULL;
    return roxml_get_parent(node);
}

static inline xml_node_t *xml_get_node_child(xml_node_t *node) {
    if (!node) return NULL;
    return roxml_get_chld(node, NULL, 0);
}

static inline xml_node_t *xml_get_node_next(xml_node_t *node) {
    if (!node) return NULL;
    return roxml_get_next_sibling(node);
}

// Fake structure members for compatibility
struct mxml_node_s {
    int type;
    mxml_value_t value;
    xml_node_t *parent;
    xml_node_t *child;
    xml_node_t *next;
};

// Override node access with compatibility functions
#define MXML_NODE_TYPE(node) xml_get_node_type(node)
#define MXML_NODE_OPAQUE(node) xml_get_node_opaque(node)
#define MXML_NODE_ELEMENT_NAME(node) xml_get_node_element_name(node)
#define MXML_NODE_PARENT(node) xml_get_node_parent(node)
#define MXML_NODE_CHILD(node) xml_get_node_child(node)
#define MXML_NODE_NEXT(node) xml_get_node_next(node)

#endif // HAVE_LIBROXML

#endif // XML_COMPAT_H
