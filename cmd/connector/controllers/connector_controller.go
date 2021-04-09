package controllers

import (
	"log"
	"time"

	"github.com/att-comdev/jarvis-connector/services"
	"github.com/att-comdev/jarvis-connector/types"
)

var (
	Connector connectorController = &ConnectorControllerImpl{
		pendingCheck:  make(chan *types.PendingChecksInfo, 5),
		pendingSubmit: make(chan *types.PendingSubmitInfo, 5),
	}
)

type connectorController interface {
	ServeCheck()
	ServeSubmit()
	PendingLoop()
}

type ConnectorControllerImpl struct {
	pendingCheck  chan *types.PendingChecksInfo
	pendingSubmit chan *types.PendingSubmitInfo
}

// ServeCheck runs the serve loop, dispatching for checks that need it.
func (controller *ConnectorControllerImpl) ServeCheck() {
	for p := range controller.pendingCheck {
		// TODO: parallelism?.
		if err := services.GerritChecker.ExecuteCheck(p); err != nil {
			log.Printf("ExecuteCheck(%v): %v", p, err)
		}
	}
}

// ServeSubmit runs the serve loop, dispatching for submissions that need it.
func (controller *ConnectorControllerImpl) ServeSubmit() {
	for p := range controller.pendingSubmit {
		// TODO: parallelism?.
		if err := services.GerritSubmitter.ExecuteSubmit(p); err != nil {
			log.Printf("ExecuteSubmit(%v): %v", p, err)
		}
	}
}

// pendingLoop periodically contacts gerrit to find new checks and submissions to
// execute. It should be executed in a goroutine.
func (controller *ConnectorControllerImpl) PendingLoop() {
	for {
		// TODO: real rate limiting.
		time.Sleep(10 * time.Second)
		pendingChecks, err := services.GerritChecker.PendingChecksByScheme(checkerScheme)
		if err == nil {
			log.Printf("Received %d Pending Checks", len(pendingChecks))
			for _, pc := range pendingChecks {
				select {
				case controller.pendingCheck <- pc:
				default:
					log.Println("too busy; dropping check.")
				}
			}
		} else {
			log.Printf("PendingChecksByScheme: %v", err)
		}

		// Handle Submissions
		pendingSubmissions, err := services.GerritSubmitter.PendingSubmit()
		if err == nil {
			log.Printf("Received %d Pending Submissions", len(pendingSubmissions))
			for _, ps := range pendingSubmissions {
				select {
				case controller.pendingSubmit <- ps:
				default:
					log.Println("too busy; dropping submission.")
				}
			}
		} else {
			log.Printf("PendingSubmit: %v", err)
		}
	}
}
