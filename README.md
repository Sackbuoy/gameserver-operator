# Gameserver-operator
## Structure
Types:
- crdEvent
- Watcher
- 

Watcher:
New(crdType: string, kubeClient: idk) -> Watcher
Watch(crdInstanceChannel: chan crdEvent)
