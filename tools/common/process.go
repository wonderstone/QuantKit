package common

type Process struct {
	progress float64

	step float64

	start float64
	end   float64

	nextStepReduceAt float64
}

func NewProcess(step, start, end float64) *Process {
	return &Process{
		progress:         start,
		step:             step,
		start:            start,
		end:              end,
		nextStepReduceAt: (start + end) / 2.0,
	}
}

func (p *Process) GetProgress() float64 {
	// 每次递增的进度
	p.progress += p.step

	// 到达上次减半时剩余量一半的时候再减半
	if p.progress >= p.nextStepReduceAt {
		p.step /= 2
		p.nextStepReduceAt = (p.end + p.progress) / 2
	}

	// 如果进度超过了end，直接设置为end
	if p.progress > p.end {
		p.progress = p.end
	}

	return p.progress
}
