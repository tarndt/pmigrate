#include <stdio.h>
#include <stdbool.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

const unsigned int runEveryXSecs = 5;
const char* defFilepath = "files_output.txt";

int main(int argc, char* argv[]) {
	char* filepath = (char*)defFilepath;
	if(argc > 1 && strlen(argv[1]) > 0) {
		filepath = argv[1];
	}	
	printf("Reading and appending to: %s\n", filepath);
	
	FILE *fout = fopen(filepath, "a");
	if(fout == NULL) {
		fprintf(stderr, "Opening: %s for writing failed!", filepath);
		return EXIT_FAILURE;
	}
	FILE* fin = fopen(filepath, "r");
	if(fin == NULL) {
		fprintf(stderr, "Opening: %s for reading failed!", filepath);
		return EXIT_FAILURE;
	}	
	
	for(unsigned long long int i = 0; true; i++) {
		//Append hello message to file
		fprintf(fout, "Hello world #%llu!\n", i);
		fflush(fout);
		if(ferror(fout) != 0) {
			fprintf(stderr, "Writing to: %s failed!", filepath);
			return EXIT_FAILURE;
		}
		
		printf("\n%s contains:\n", filepath);	
		//Read messages in fie
		rewind(fin);
		for(int c = fgetc(fin); EOF != c; c = fgetc(fin)) {
			putc(c, stdout);
		}
		if(ferror(fin) != 0) {
			fprintf(stderr, "Reading from: %s failed!", filepath);
			return EXIT_FAILURE;
		}		
		
		sleep(runEveryXSecs);
	}
	
	fclose(fout);
	fclose(fin);
	return EXIT_SUCCESS;
}
