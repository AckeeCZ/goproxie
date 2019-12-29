package main

func main() {
	// TODO: Fetch GCP Projects for the User
	// TODO: Prompt user to select a GCP Project {=GCPPROJECT}
	//	I suggest using Node.js's Inquirer go equivalent Survey https://github.com/AlecAivazis/survey
	// TODO: Prompt user for a type of proxy (CloudSQL, VM, Pod) {=PROXYTYPE}
	// 	Pod and VM should be fairly easy. CloudSQL probably won't have any SDK support
	//	and user would must have it installed. Goproxie would then call the installed binary.
	// For PROXYTYPE=Pod
	//	TODO: Fetch Clusters for the selected GCPPROJECT
	//	TODO: Prompt user to select the Cluster {=CLUSTER}
	//	TODO: Fetch Pods for the selected GCPPROJECT and CLUSTER
	//	TODO: Prompt user to select the Pod {=POD}
	//	TODO: Prompt user for local port number
	// 	TODO (LOW): Prefill the port with the most appropriate and free port on local machine
	//		Current Node.js/proxie logic is: If it's a VM and contains `mongo` in the name,
	//		use MongoDB's native port 27017 as the starting point for our guess,
	//		if it's CloudSQL proxy type then 3306,
	//		if Pod and name contains `mongo` in the name then 27017,
	//		3000 otherwise. Then look for first free port by increasing the number - if 3306 is full, try 3307 etc.
	//		Original logic can be foud here: https://github.com/AckeeCZ/be-scripts/blob/master/src/lib/proxie.ts#L167
	//	TODO: Prompt user for remote port number
	//	TODO (LOW): Prefill by the logic above. Remote service almost always uses the default port.
	//	TODO: Create a kubectl port-forward for the given GCPPROJECT, CLUSTER, POD and ports.
	// For PROXYTYPE=VM
	//	TODO: Complete this guide.
	// For PROXYTYPE=CloudSQL
	//	TODO: Complete this guide.
}
