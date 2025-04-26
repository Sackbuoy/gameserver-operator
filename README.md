# Gameserver-operator
## Structure
Types:
- crdEvent
- Watcher
- 

Watcher:
New(crdType: string, kubeClient: idk) -> Watcher
Watch(crdInstanceChannel: chan crdEvent)

TODO:
1. Update functionality
2. optimize
3. add ctx everywhere
4. formatting
