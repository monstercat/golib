package logger

const (
	TypeMetric = "metric"
)

// MetricEntry is used to create a specific type of log for use in generating metrics through a cloud logging platform.
// To use it, create new functions, for example:
//
// NewXXXLogMetric(value int) LogMetric {
//    return LogMetric{
//      Type: "metric",
//      ....
//    }
// }
type MetricEntry struct {
	// Type is always "metric"
	Type string

	// Name is the name of the metric.
	Name string

	// Namespace should resolve to [connect, http] in this case
	Namespace []string

	// Value is the specific value of the metric
	Value int

	// Description of the metric, if any.
	Description string

	// Unit is the unit that the metric is counted in. (e.g., count, ms)
	Unit string

	// Any other context information.
	Context interface{}
}

// SetContext is used to implement the ContextualEntry interface
func (l *MetricEntry) SetContext(context interface{}) {
	l.Context = context
}

func (l *MetricEntry) GetContext() interface{} {
	return l.Context
}

func (l *MetricEntry) GetLog() interface{} {
	return map[string]interface{}{
		"Type":        l.Type,
		"Name":        l.Name,
		"Namespace":   l.Namespace,
		"Value":       l.Value,
		"Description": l.Description,
		"Unit":        l.Unit,
	}
}

func (l *MetricEntry) SetDefault() {
	l.Type = TypeMetric
}
