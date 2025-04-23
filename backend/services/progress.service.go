package services

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type (
	IProgressService interface {
		SetMax(max int)
		Step()
		StepFile(file string)
		Start()
		Stop()
	}

	ProgressService struct {
		ctx         context.Context
		Total       int    `json:"total"`
		Processed   int    `json:"processed"`
		Percentage  int    `json:"percentage"`
		CurrentFile string `json:"file"`
	}
)

func NewProgressService(ctx context.Context) IProgressService {
	return &ProgressService{
		ctx:         ctx,
		Total:       0,
		Processed:   0,
		Percentage:  0,
		CurrentFile: "",
	}
}

func (p *ProgressService) SetMax(max int) {
	p.Total = max
}

func (p *ProgressService) Step() {
	p.Processed++
	p.Percentage = (p.Processed * 100) / p.Total
	p.sendProgress()
}

func (p *ProgressService) StepFile(file string) {
	p.CurrentFile = file

	p.Processed++
	p.Percentage = (p.Processed * 100) / p.Total
	p.sendProgress()
}

func (p *ProgressService) Start() {
	p.sendProgress()
	p.showProgress()
}

func (p *ProgressService) Stop() {
	p.restart()
}

func (p *ProgressService) Restart() {
	p.Stop()
	p.Start()
}

func (p *ProgressService) SendProgress(progress ProgressService) {
	runtime.EventsEmit(p.ctx, "Progress", progress)
}

func (p *ProgressService) setProcessed(processed int) {
	p.Processed = processed
}

func (p *ProgressService) setPercentage(percentage int) {
	p.Percentage = percentage
}

func (p *ProgressService) sendProgress() {
	runtime.EventsEmit(p.ctx, "Progress", p)
}

func (p *ProgressService) showProgress() {
	runtime.EventsEmit(p.ctx, "ShowProgress", true)
}

func (p *ProgressService) restart() {
	p.SetMax(0)
	p.setProcessed(0)
	p.setPercentage(0)
	p.CurrentFile = ""
}
