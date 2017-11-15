package metrics

import (
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/fatih/set.v0"
	"math/rand"
	"strings"
	"time"
)

// Defaults
const (
	IntRatio                 = 90
	FloatRatio               = 10
	TagSize                  = 6
	MetricPerMetricNameRatio = 2
	TagsPerMetric            = 4
	MaxNumValuePerTag        = 10
	MetricNameSize           = 20
	BufferSize               = 100000
	longForm = "2006-01-02 15:04:05"
	shortForm = "2006-01-02"
)

func ParseTimeStamp(s string) time.Time {
	// try with the long form first
	t, err := time.Parse(longForm,s)
	if err != nil {
		t, err = time.Parse(shortForm,s)
		if err != nil {
			panic(err)
		}
		return t
	}
	return t
}

type ConfigSet struct {
	IntRatio                 int
	FloatRatio               int
	TagSize                  int
	MetricPerMetricNameRatio int
	TagsPerMetric            int
	MaxNumValuePerTag        int
	MetricNameSize           int
	BufferSize               int
	NumMetrics               int
	NumTags                  int
	MandatoryTags            map[string]int
	Start                    string
	End                      string
	Step                     int64
}

func NewConfigSet() *ConfigSet {
	// Setting default values
	var c ConfigSet
	c.IntRatio = IntRatio
	c.FloatRatio = FloatRatio
	c.TagSize = TagSize
	c.MetricPerMetricNameRatio = MetricPerMetricNameRatio
	c.TagsPerMetric = TagsPerMetric
	c.MaxNumValuePerTag = MaxNumValuePerTag
	c.MetricNameSize = MetricNameSize
	c.BufferSize = BufferSize
	return &c
}

type MetricInt struct {
	name  string
	tags  string
	value int64
}

func (m *MetricInt) String() string {
	return fmt.Sprintf("%s,%s value=%d", m.name, m.tags, m.value)
}

func (m *MetricInt) Change(t int64) {
	m.value += rand.Int63n(10) - 4
}

func NewMetricInt(name string, tags string) *MetricInt {
	return &MetricInt{name, tags, rand.Int63n(1000)}
}

type MetricFloat struct {
	name  string
	tags  string
	value float64
}

func (m *MetricFloat) String() string {
	return fmt.Sprintf("%s,%s value=%.4f", m.name, m.tags, m.value)
}

func (m *MetricFloat) Change(t int64) {
	m.value += rand.NormFloat64() * 10
}

func NewMetricFloat(name string, tags string) *MetricFloat {
	return &MetricFloat{name, tags, rand.NormFloat64() * 100}
}

type MetricBool struct {
	name  string
	tags  string
	value bool
}

func (m *MetricBool) String() string {
	return fmt.Sprintf("%s,%s value=%t", m.name, m.tags, m.value)
}

func (m *MetricBool) Change(t int64) {
	m.value = rand.Intn(20) < 16
}

func NewMetricBool(name string, tags string) *MetricBool {
	return &MetricBool{name, tags, true}
}

type Metric interface {
	String() string
	Change(t int64)
}

type TagsFactory map[string][]string

func NewTagsFactoryFromList(tagList map[string]int) *TagsFactory {
	tf := make(TagsFactory, len(tagList))
	for k, v := range tagList {
		tf[k] = make([]string, v)
		for i := 0; i < v; i++ {
			tf[k][i] = uuid.New().String()[:TagSize]
		}
	}
	return &tf
}

func NewTagsFactoryFromNum(numTags int) *TagsFactory {
	tf := make(TagsFactory, numTags)
	for t := 0; t < numTags; t++ {
		tagName := "tag_" + uuid.New().String()[:TagSize]
		for k := 0; k < MaxNumValuePerTag; k++ {
			tf[tagName] = append(tf[tagName], string('_')+uuid.New().String()[:TagSize])
		}
	}
	return &tf
}

func (tf TagsFactory) KVAllTags() []string {
	arr := make([]string, len(tf))
	var count int
	for s, v := range tf {
		// roulette
		ind := rand.Intn(len(v))
		arr[count] = fmt.Sprintf("%s=%s", s, v[ind])
		count++
	}
	return arr
}

func (tf TagsFactory) KVSomeTags(numTags int) []string {
	var arr []string
	// selects the tags
	s := set.New(rand.Intn(len(tf)))
	for s.Size() < numTags {
		s.Add(rand.Intn(len(tf)))
	}
	var count int
	for k, v := range tf {
		// roulette
		if s.Has(count) {
			// roulette
			ind := rand.Intn(len(v))
			arr = append(arr, fmt.Sprintf("%s=%s", k, v[ind]))
		}
		count++
	}
	return arr
}

type MetricFactory struct {
	metricList   []Metric
	timestamp    int64
	step         int64
	endTimestamp int64
	Output       chan string
	Stop         chan bool
}

func NewMetricFactory(c *ConfigSet) *MetricFactory {

	mandatoryTagMap := NewTagsFactoryFromList(c.MandatoryTags)
	optionalTag := c.TagsPerMetric - len(c.MandatoryTags)
	optionalTagMap := NewTagsFactoryFromNum(optionalTag)

	ml := make([]Metric, c.NumMetrics)

	// calculating ratios
	// metric < limit_int: metric will be INT
	// limit_int <= metric < limit_float: metric will be FLOAT
	// limit_float < metric: metric will be BOOL
	var limitInt = c.NumMetrics * c.IntRatio / 100
	var limitFloat = c.NumMetrics * (c.FloatRatio + c.IntRatio) / 100

	// calculating how many metrics per metric name (min 1)
	var numPerMetricName = c.NumMetrics * c.MetricPerMetricNameRatio / 100
	if numPerMetricName == 0 {
		numPerMetricName = 1
	}

	var i = 0
	for {
		if i >= c.NumMetrics {
			break
		}
		if i < limitInt {
			metricName := "int." + uuid.New().String()[:c.MetricNameSize]
			for m := 0; m < numPerMetricName && i < c.NumMetrics; m++ {
				mt := strings.Join(append(mandatoryTagMap.KVAllTags(), optionalTagMap.KVSomeTags(optionalTag)...), ",")
				ml[i] = NewMetricInt(metricName, mt)
				i++
			}
		} else if i < limitFloat {
			metricName := "float." + uuid.New().String()[:c.MetricNameSize]
			for m := 0; m < numPerMetricName && i < c.NumMetrics; m++ {
				mt := strings.Join(append(mandatoryTagMap.KVAllTags(), optionalTagMap.KVSomeTags(optionalTag)...), ",")
				ml[i] = NewMetricFloat(metricName, mt)
				i++
			}
		} else {
			metricName := "bool." + uuid.New().String()[:c.MetricNameSize]
			for m := 0; m < numPerMetricName && i < c.NumMetrics; m++ {
				mt := strings.Join(append(mandatoryTagMap.KVAllTags(), optionalTagMap.KVSomeTags(optionalTag)...), ",")
				ml[i] = NewMetricBool(metricName, mt)
				i++
			}
		}
	}

	// building the metric factory
	mf := &MetricFactory{}
	mf.metricList = ml
	// parsing timestamps
	mf.timestamp = ParseTimeStamp(c.Start).UnixNano()
	mf.endTimestamp = ParseTimeStamp(c.End).UnixNano()
	mf.step = c.Step
	mf.Output = make(chan string, c.BufferSize)
	mf.Stop = make(chan bool)

	return mf
}

func (mf *MetricFactory) Produce() {
	// current round:
	for {
		select {
		case <-mf.Stop:
			close(mf.Output)
			return
		default:
			for _, metric := range mf.metricList {
				mf.Output <- fmt.Sprintf("%s %d", metric.String(), mf.timestamp)
				metric.Change(mf.step)
			}
		}
		mf.timestamp += mf.step
		if mf.timestamp >= mf.endTimestamp {
			close(mf.Output)
			return
		}
	}
}
