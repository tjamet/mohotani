package kubernetes

import (
	"time"

	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/informers/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type DomainLister struct {
	informer     v1beta1.IngressInformer
	hosts        []string
	ingressHosts map[string][]string
	class        string
}

func NewDomainLister(c kubernetes.Interface, class string) *DomainLister {
	return &DomainLister{
		informers.NewSharedInformerFactory(c, 30*time.Second).Extensions().V1beta1().Ingresses(),
		[]string{},
		map[string][]string{},
		class,
	}
}

func (dl *DomainLister) Listen(out chan []string) {
	informer := dl.informer.Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: dl.handle(out),
		UpdateFunc: func(old, obj interface{}) {
			dl.handle(out)(obj)
		},
		DeleteFunc: func(obj interface{}) {
			if ing, ok := obj.(*extensionsv1beta1.Ingress); ok {
				delete(dl.ingressHosts, ing.GetSelfLink())
				dl.notify(out)
			}
		},
	})
	informer.Run(make(chan struct{}))
}

func (dl *DomainLister) handle(out chan []string) func(interface{}) {
	return func(obj interface{}) {
		if ing, ok := obj.(*extensionsv1beta1.Ingress); ok {
			if class, ok := ing.Annotations["kubernetes.io/ingress.class"]; ok && class != dl.class {
				return
			}
			if enabled, ok := ing.Annotations["mohotani.io/enable"]; !ok || enabled != "true" {
				return
			}
			hosts := []string{}
			for _, rule := range ing.Spec.Rules {
				hosts = append(hosts, rule.Host)
			}
			dl.ingressHosts[ing.GetSelfLink()] = hosts
			dl.notify(out)

		}
	}
}

func (dl *DomainLister) notify(out chan []string) {
	hosts := map[string]interface{}{}
	for _, ingHosts := range dl.ingressHosts {
		for _, h := range ingHosts {
			hosts[h] = nil
		}
	}
	if len(hosts) > 0 {
		uniqueHosts := []string{}
		for k := range hosts {
			uniqueHosts = append(uniqueHosts, k)
		}
		out <- uniqueHosts
	}
}
