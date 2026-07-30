#include <stdlib.h>
#include <stdio.h>
#include "object.h"

// The shortest representation that reads back as the very same value, the
// way Go prints a float64. "%f" would round 3.14159265 down to 3.141593 and
// pad every whole number with six zeroes.
char *float_str(struct object o) {
	char *str = calloc(40, sizeof(char));

	for (int prec = 15; prec <= 17; prec++) {
		snprintf(str, 40, "%.*g", prec, o.data.f);
		if (strtod(str, NULL) == o.data.f) {
			break;
		}
	}

	return str;
}

struct object new_float_obj(double val) {
	return (struct object) {
		.data.f = val,
		.type = obj_float,
	};
}
