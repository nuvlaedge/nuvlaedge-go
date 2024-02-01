package jobProcessor

type Action interface {
	DoYourThing()
}

type DeploymentStart struct {
}

func (d *DeploymentStart) DoYourThing() {

}
