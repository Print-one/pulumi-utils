package main

import (
	"context"
	"fmt"
	"os"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

func main() {
	provider, err := infer.NewProviderBuilder().
		WithResources(
			infer.Resource(Queue{}),
		).
		WithNamespace("print-one").
		Build()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}

	err = provider.Run(context.Background(), "pulumi-utils", "0.1.0")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}

type Queue struct{}

func (f *Queue) Annotate(a infer.Annotator) {
	a.Describe(&f, "A pulumi resource that returns the value of the previous deploy")
}

type QueueArgs struct {
	In      string `pulumi:"in"`
	Default string `pulumi:"default,optional"`
}

func (f *QueueArgs) Annotate(a infer.Annotator) {
	a.Describe(&f.In, "The content of the queue")
	a.Describe(&f.Default, "The value of the first create defaults to `in`")
}

type QueueState struct {
	In  string `pulumi:"in"`
	Out string `pulumi:"out"`
}

func (q *QueueState) Annotate(a infer.Annotator) {
	a.Describe(&q.In, "The current value of the queue")
	a.Describe(&q.Out, "The previous value of the queue")
}

func (q *QueueArgs) GetCreateOut() string {
	if q.Default != "" {
		return q.Default
	}
	return q.In
}

func (Queue) Create(ctx context.Context, req infer.CreateRequest[QueueArgs]) (resp infer.CreateResponse[QueueState], err error) {
	return infer.CreateResponse[QueueState]{
		ID: req.Name,
		Output: QueueState{
			In:  req.Inputs.In,
			Out: req.Inputs.GetCreateOut(),
		}}, nil
}

func (Queue) Update(ctx context.Context, req infer.UpdateRequest[QueueArgs, QueueState]) (infer.UpdateResponse[QueueState], error) {
	return infer.UpdateResponse[QueueState]{
		Output: QueueState{
			In:  req.Inputs.In,
			Out: req.State.In,
		},
	}, nil
}

func (Queue) Diff(ctx context.Context, req infer.DiffRequest[QueueArgs, QueueState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.Inputs.In != req.State.Out || req.Inputs.In != req.State.In {
		diff["in"] = p.PropertyDiff{Kind: p.Update}
	}
	return infer.DiffResponse{
		HasChanges:   len(diff) > 0,
		DetailedDiff: diff,
	}, nil
}

func (Queue) WireDependencies(f infer.FieldSelector, args *QueueArgs, state *QueueState) {
	f.OutputField(&state.Out).DependsOn(f.InputField(&args.In))
	f.OutputField(&state.In).DependsOn(f.InputField(&args.In))
}
