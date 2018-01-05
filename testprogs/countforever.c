#include <stdio.h>
#include <stdbool.h>
#include <stdlib.h>

int main() {
	for(unsigned long long int i = 0; true; i++) {
		printf("%llu\n", i);
		fflush(stdout);
	}
	return EXIT_SUCCESS;
}
