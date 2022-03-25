package listens

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"sync"
)

var EvenMap *EventMapStrct

type EventMapStrct struct {
	data sync.Map
}

func init() {
	EvenMap = &EventMapStrct{}
}

func (e *EventMapStrct) GetMessage(ns, kind, name string) string {
	key := fmt.Sprintf("%s_%s_%s", ns, kind, name)
	if v, ok := e.data.Load(key); ok {
		return v.(*corev1.Event).Message
	}
	fmt.Println("没找" + key + "的到事件")
	return ""
}

type EvensHandler struct {
}

func (d *EvensHandler) StoreData(obj interface{}, isDelete bool) {
	if event, ok := obj.(*corev1.Event); ok {
		key := fmt.Sprintf("%s_%s_%s", event.InvolvedObject.Namespace, event.InvolvedObject.Kind, event.InvolvedObject.Name)
		fmt.Sprintf("保存%s的事件%s", key, event.Message)
		if !isDelete {
			EvenMap.data.Store(key, event)
		} else {
			EvenMap.data.Delete(key)
		}
	}
}

func (d *EvensHandler) OnAdd(obj interface{}) {
	d.StoreData(obj, false)
}

func (d *EvensHandler) OnUpdate(oldObj, newObj interface{}) {
	d.StoreData(newObj, false)
}

func (d *EvensHandler) OnDelete(obj interface{}) {
	d.StoreData(obj, true)
}
