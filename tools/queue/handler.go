package queue

import (
	"github.com/golang/glog"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// QueueingEventHandler queues the key for the object on add and update events
type QueueingEventHandler struct {
	queue         workqueue.RateLimitingInterface
	enqueueAdd    bool
	enqueueUpdate func(oldObj, newObj interface{}) bool
	enqueueDelete bool
}

var _ cache.ResourceEventHandler = &QueueingEventHandler{}

func DefaultEventHandler(queue workqueue.RateLimitingInterface) cache.ResourceEventHandler {
	return &QueueingEventHandler{
		queue:         queue,
		enqueueAdd:    true,
		enqueueUpdate: nil,
		enqueueDelete: true,
	}
}

func NewEventHandler(queue workqueue.RateLimitingInterface, enqueueUpdate func(oldObj, newObj interface{}) bool) cache.ResourceEventHandler {
	return &QueueingEventHandler{
		queue:         queue,
		enqueueAdd:    true,
		enqueueUpdate: enqueueUpdate,
		enqueueDelete: true,
	}
}

func NewDeleteHandler(queue workqueue.RateLimitingInterface) cache.ResourceEventHandler {
	return &QueueingEventHandler{
		queue:         queue,
		enqueueAdd:    false,
		enqueueUpdate: func(oldObj, newObj interface{}) bool { return false },
		enqueueDelete: true,
	}
}

func (h *QueueingEventHandler) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		glog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		return
	}
	h.queue.Add(key)
}

func (h *QueueingEventHandler) OnAdd(obj interface{}) {
	glog.V(6).Infof("Add event for %+v\n", obj)
	if h.enqueueAdd {
		h.enqueue(obj)
	}
}

func (h *QueueingEventHandler) OnUpdate(oldObj, newObj interface{}) {
	glog.V(6).Infof("Update event for %+v\n", newObj)
	if h.enqueueUpdate == nil || h.enqueueUpdate(oldObj, newObj) {
		h.enqueue(newObj)
	}
}

func (h *QueueingEventHandler) OnDelete(obj interface{}) {
	glog.V(6).Infof("Delete event for %+v\n", obj)
	if h.enqueueDelete {
		h.enqueue(obj)
	}
}
