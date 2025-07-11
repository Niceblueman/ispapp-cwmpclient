/*
 * easycwmp.c - Easy CWMP client compatibility functions
 *
 * This file provides compatibility functions for easycwmp library
 * that may be referenced by other parts of the ispappd project.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <uci.h>
#include "config.h"
#include "ispappcwmp.h"

/* UCI context for ispapp compatibility */
static struct uci_context *ISPAPPCWMP_uci_ctx = NULL;

/* Initialize UCI context */
int ISPAPPCWMP_uci_init(void)
{
    if (ISPAPPCWMP_uci_ctx)
        uci_free_context(ISPAPPCWMP_uci_ctx);
    
    ISPAPPCWMP_uci_ctx = uci_alloc_context();
    if (!ISPAPPCWMP_uci_ctx)
        return -1;
    
    return 0;
}

/* Free UCI context */
void ISPAPPCWMP_uci_fini(void)
{
    if (ISPAPPCWMP_uci_ctx) {
        uci_free_context(ISPAPPCWMP_uci_ctx);
        ISPAPPCWMP_uci_ctx = NULL;
    }
}

/* Get UCI value */
char *ISPAPPCWMP_uci_get_value(const char *package, const char *section, const char *option)
{
    struct uci_ptr ptr;
    char *value = NULL;
    char lookup_str[256];
    
    if (!ISPAPPCWMP_uci_ctx)
        return strdup("");
    
    snprintf(lookup_str, sizeof(lookup_str), "%s.%s.%s", package, section, option);
    
    if (uci_lookup_ptr(ISPAPPCWMP_uci_ctx, &ptr, lookup_str, true) != UCI_OK)
        return strdup("");
    
    if (ptr.o && ptr.o->v.string)
        value = strdup(ptr.o->v.string);
    else
        value = strdup("");
    
    return value;
}

/* Set UCI value */
int ISPAPPCWMP_uci_set_value(const char *package, const char *section, const char *option, const char *value)
{
    struct uci_ptr ptr;
    char lookup_str[256];
    
    if (!ISPAPPCWMP_uci_ctx)
        return -1;
    
    if (option) {
        snprintf(lookup_str, sizeof(lookup_str), "%s.%s.%s", package, section, option);
    } else {
        snprintf(lookup_str, sizeof(lookup_str), "%s.%s", package, section);
    }
    
    if (uci_lookup_ptr(ISPAPPCWMP_uci_ctx, &ptr, lookup_str, true) != UCI_OK)
        return -1;
    
    if (option) {
        ptr.value = value;
        if (uci_set(ISPAPPCWMP_uci_ctx, &ptr) != UCI_OK)
            return -1;
    } else {
        /* Creating section */
        if (!ptr.s) {
            ptr.value = value;
            if (uci_set(ISPAPPCWMP_uci_ctx, &ptr) != UCI_OK)
                return -1;
        }
    }
    
    return 0;
}

/* Commit UCI changes */
int ISPAPPCWMP_uci_commit(void)
{
    if (!ISPAPPCWMP_uci_ctx)
        return -1;
    
    return uci_commit(ISPAPPCWMP_uci_ctx, NULL, false);
}

/* Compatibility function for old naming */
int ispapp_uci_init(void)
{
    return ISPAPPCWMP_uci_init();
}

void ispapp_uci_fini(void)
{
    ISPAPPCWMP_uci_fini();
}

char *ispapp_uci_get_value(const char *package, const char *section, const char *option)
{
    return ISPAPPCWMP_uci_get_value(package, section, option);
}

int ispapp_uci_set_value(const char *package, const char *section, const char *option, const char *value)
{
    return ISPAPPCWMP_uci_set_value(package, section, option, value);
}

int ispapp_uci_commit(void)
{
    return ISPAPPCWMP_uci_commit();
}
