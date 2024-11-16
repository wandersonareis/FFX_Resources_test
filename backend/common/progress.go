package common

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type IProgress interface {
	SetMax(max int)
	SetProcessed(processed int)
	SetPercentage(percentage int)
	SendProgress(progress Progress)
	Step()
	Start()
	Stop()
}

type Progress struct {
	ctx        context.Context
	Total      int `json:"total"`
	Processed  int `json:"processed"`
	Percentage int `json:"percentage"`
}

func NewProgress(ctx context.Context) IProgress {
	return &Progress{
		ctx: ctx,
	}
}

func (p *Progress) SetMax(max int) {
	p.Total = max
}

func (p *Progress) SetProcessed(processed int) {
	p.Processed = processed
}

func (p *Progress) SetPercentage(percentage int) {
	p.Percentage = percentage
}

func (p *Progress) Step() {
	p.Processed++
	p.Percentage = (p.Processed * 100) / p.Total
	p.sendProgress()
}

func (p *Progress) Start() {
	p.sendProgress()
	runtime.EventsEmit(p.ctx, "ShowProgress", true)
}

func (p *Progress) Stop() {
	p.restart()
}

func (p *Progress) Restart() {
	p.Stop()
	p.Start()
}

func (p *Progress) SendProgress(progress Progress) {
	runtime.EventsEmit(p.ctx, "Progress", progress)
}

func (p *Progress) sendProgress() {
	runtime.EventsEmit(p.ctx, "Progress", p)
}

func (p *Progress) restart() {
	p.SetMax(0)
	p.SetProcessed(0)
	p.SetPercentage(0)
}