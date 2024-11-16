#include "vm.h"

inline struct heap new_heap(int64_t treshold) {
    return (struct heap) {
        .root = NULL,
        .len = 0,
        .treshold = treshold,
    };
}

inline void heap_add(struct heap *h, struct object obj) {
    struct heap_node *node = malloc(sizeof(struct heap_node));
    node->next = h->root;
    node->obj = obj;

    h->root = node;
    h->treshold *= (++h->len >= h->treshold) + 1;
}

inline void heap_dispose(struct heap *h) {
    for (struct heap_node *n = h->root; n != NULL;) {
        struct heap_node *tmp = n->next;
        struct object obj = n->obj;

        // We delete the object only if there are no more references to it.
        if (dec_refcnt(obj.gcdata) == 0) {
            free_obj(obj);
        }
        free(n);
        n = tmp;
    }
    h->root = NULL;
    h->len = 0;
    h->treshold = 1024;
}
