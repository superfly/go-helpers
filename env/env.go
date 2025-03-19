// Exposes environment variables that should be populated on Fly.io machines as
// package variables for convenience.
package fly

import "os"

var (
	// App name: Each app running on Fly.io has a unique app name. This
	// identifies the app for the user, and can also identify its Machines on
	// their IPv6 private network, using internal DNS. For example,
	// syd.$FLY_APP_NAME.internal can refer to the app’s Machines running in the
	// syd region. Read more about 6PN Naming in the Private Networking docs.
	AppName = os.Getenv("FLY_APP_NAME")

	// Machine ID: The Machine’s unique ID on the Fly.io platform. This is the
	// ID you use to target the Machine using flyctl and the Machines API. You
	// can also see it in the log messages emitted by the Machine.
	MachineID = os.Getenv("FLY_MACHINE_ID")

	// Allocation ID: Same as the FLY_MACHINE_ID.
	AllocID = os.Getenv("FLY_ALLOC_ID")

	//  Region name: The three-letter name of the region the Machine is running
	//  in. Details of current regions are listed in the Regions page. As an
	//  example, “ams” is the region name for Amsterdam.
	//
	// Not to be confused with the HTTP header Fly-Region, which is where the
	// connection was accepted from.
	Region = os.Getenv("FLY_REGION")

	// IPV6 public IP: The full public outbound IPV6 address for this Machine.
	// Read more in the Network Services section.
	PublicIP = os.Getenv("FLY_PUBLIC_IP")

	//  Docker image reference: The name of the Docker image used to create the
	//  Machine.
	//  registry.fly.io/my-app-name:deployment-01H9RK9EYO9PGNBYAKGXSHV0PH is an
	//  example of the Docker Image Reference’s format.
	//
	// Useful if your app needs to launch Machine instances of itself to scale
	// background workers to zero and back, as in Rails Background Jobs with Fly
	// Machines.
	ImageRef = os.Getenv("FLY_IMAGE_REF")

	// Machine configuration version: A version identifier associated with a
	// specific Machine configuration. When you update a Machine’s configuration
	// (including when you update its Docker image), it gets a new
	// FLY_MACHINE_VERSION. Changing the Machine’s metadata using the metadata
	// endpoint of the Machines API doesn’t trigger a new version. You can also
	// find this value under the name Instance ID in the output of fly machine
	// status.
	MachineVersion = os.Getenv("FLY_MACHINE_VERSION")

	// Private IPv6 address: The IPv6 address of the Machine on its 6PN private network.
	PrivateIP = os.Getenv("FLY_PRIVATE_IP")

	// Process group: The Fly Launch process group associated with the Machine, if any.
	ProcessGroup = os.Getenv("FLY_PROCESS_GROUP")

	// Machine memory: The memory allocated to the Machine, in MB. It’s the same
	// value you’ll find under https://fly.io/dashboard/personal/machines and VM
	// Memory in the output of fly machine status. Learn more about Machine
	// sizing.
	VMMemoryMB = os.Getenv("FLY_VM_MEMORY_MB")

	// Primary region: This is set in your fly.toml or with the --region flag
	// during deploys. Learn more about configuring the primary region.
	PrimaryRegion = os.Getenv("FLY_PRIMARY_REGION")
)
