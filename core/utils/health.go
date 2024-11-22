package utils

import (
	"fmt"
	"github.com/apex/log"
	"net/http"
)

func HandleHealthFn(healthFn func() bool, healthPort int) {
	addr := fmt.Sprintf(":%d", healthPort)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Errorf("health check server failed: %v", err)
		}
	}()
	log.Infof("health check server started on port %d", healthPort)
}
