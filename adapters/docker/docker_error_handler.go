package docker

import "log"

func handleUnexpectedError(err error) {
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}
}

func handleUnfoundServiceName(serviceName string) {
	log.Fatalf("service %s not found", serviceName)
}
