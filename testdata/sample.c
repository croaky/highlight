#include <stdlib.h>
#include <string.h>

#include "ring.h"

#define CAP 64
#define MIN(a, b) ((a) < (b) ? (a) : (b))

/* A fixed ring of bytes, which is every buffer this needed and one
   fewer allocation than a growing one. */
typedef struct {
	uint8_t len;
	uint16_t head;
	char buf[CAP];
} Ring;

static const char *kName = "ring";
static const char kSep = '\n';

Ring *ring_new(void) {
	Ring *r = calloc(1, sizeof(Ring));
	if (r == NULL) {
		return NULL; // the caller decides what a full heap means
	}
	r->head = 0;
	return r;
}

void ring_free(Ring *r) { free(r); }

bool ring_push(Ring *r, char c) {
	if (r->len >= CAP) {
		return false;
	}
	r->buf[(r->head + r->len) % CAP] = c;
	r->len++;
	return true;
}

/* Copies out at most n bytes and returns how many, so a short read is
   a number rather than an error. */
size_t ring_read(Ring *r, char *dst, size_t n) {
	size_t got = MIN(n, (size_t)r->len);
	for (size_t i = 0; i < got; i++) {
		dst[i] = r->buf[(r->head + i) % CAP];
	}
	r->head = (r->head + got) % CAP;
	r->len -= (uint8_t)got;
	return got;
}

int ring_flush(Ring *r, char *dst, size_t n) {
	size_t got = ring_read(r, dst, n);
	if (got == 0) {
		return -1;
	}
	switch (dst[got - 1]) {
	case '\n':
	case '\r':
		return (int)got - 1;
	default:
		return (int)got;
	}
}

double ring_load(const Ring *r) { return (double)r->len / 64.0; }

#ifdef RING_DEBUG
void ring_dump(const Ring *r) {
	char out[CAP + 1];
	memcpy(out, r->buf, 0x40);
	out[CAP] = '\0';
	fprintf(stderr, "%s: %u/%d %c\n", kName, r->len, CAP, kSep);
}
#endif
