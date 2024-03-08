package co2processor

import (
	"context"
	"fmt"

	"github.com/giantswarm/cloud-carbon/pkg/footprint"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type spanProcessor struct {
	config Config
}

func newSpanProcessor() (*spanProcessor, error) {
	sp := &spanProcessor{}
	return sp, nil
}

func (sp *spanProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	// Process each span in the trace data.
	rs := td.ResourceSpans()
	for i := 0; i < rs.Len(); i++ {
		rs := rs.At(i)
		// Extract the AWS instance ID and region from the resources.
		instanceID, region := sp.extractAWSInfo(rs)
		// Process each instrumentation library in the resource spans.
		ils := rs.ScopeSpans()
		for j := 0; j < ils.Len(); j++ {
			ils := ils.At(j)
			// Process each span in the instrumentation library.
			spans := ils.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				// Pass the AWS instance ID and region to the processSpan function.
				sp.processSpan(span, instanceID, region)
			}
		}
	}
	return td, nil
}

func (sp *spanProcessor) processSpan(span ptrace.Span, instanceID string, region string) {
	startTime := span.StartTimestamp()
	endTime := span.EndTimestamp()
	duration := endTime.AsTime().Sub(startTime.AsTime())
	footprint, err := footprint.AWS(region, instanceID, duration)
	if err != nil {
		// Log the error.
		zap.S().Errorf("error calculating AWS footprint: %s", err)
		return
	}
	span.Attributes().PutStr("co2.footprint", fmt.Sprintf("%.2f", footprint))
}

func (sp *spanProcessor) extractAWSInfo(rs ptrace.ResourceSpans) (string, string) {
	instanceID := ""
	region := ""
	// Iterate over the attributes of the resource.
	attrs := rs.Resource().Attributes()
	attrs.Range(func(k string, v pcommon.Value) bool {
		// Check if the attribute key matches the instance ID key.
		if k == "aws.ec2.instance.id" {
			instanceID = v.AsString()
		}
		// Check if the attribute key matches the region key.
		if k == "aws.region" {
			region = v.AsString()
		}
		return true
	})

	return instanceID, region
}
