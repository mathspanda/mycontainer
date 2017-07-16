package nsenter

/*
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

__attribute__((constructor)) void enter_namespace(void) {
    char *mycontainer_pid;
    mycontainer_pid = getenv("mycontainer_pid");
    if (!mycontainer_pid) {
        return;
    }

    char *mycontainer_cmd;
    mycontainer_cmd = getenv("mycontainer_cmd");
    if (!mycontainer_cmd) {
    	return;
    }

    int i;
    char nspath[1024];
    char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };

    for (i = 0; i < 5; i++) {
    	sprintf(nspath, "/proc/%s/ns/%s", mycontainer_pid, namespaces[i]);
    	int fd = open(nspath, O_RDONLY);
    	if (setns(fd, 0) == -1) {
    	    fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
    	}
    	close(fd);
    }

    int res = system(mycontainer_cmd);
    exit(0);
    return;
}
*/
import "C"
