package actions

//type Update struct {
//	executor executors.EngineUpdater
//
//	client *nuvla.NuvlaClient
//
//	// To stop this NuvlaEdge, it accepts SIGURS1 signal
//	jobPayload *types.UpdateJobPayload
//}
//
//// extractResources expects a list of TWO maps, each containing a single key-value pair (href:<value>).
//// One should contain the related nuvlabox and the other the related release.
//// It returns the target release first and then the nuvlabox.
//func extractResources(res []struct {
//	Href string `json:"href"`
//}) (string, string, error) {
//	var rel string
//	var box string
//	for _, r := range res {
//		if r.Href == "" {
//			continue
//		}
//		if strings.Contains(r.Href, "release") {
//			rel = r.Href
//		}
//		if strings.Contains(r.Href, "nuvlabox") {
//			box = r.Href
//		}
//	}
//	if rel == "" {
//		return "", "", errors.New("could not extract release resource")
//	}
//	return rel, box, nil
//}
//
//func GetReleaseResource(client *nuvla.NuvlaClient, relId string) (*types.NuvlaEdgeReleaseResource, error) {
//	res, err := client.Get(relId, nil)
//	if err != nil {
//		return nil, err
//	}
//	if res == nil {
//		return nil, errors.New("could not retrieve release resource")
//	}
//	return types.NewReleaseResourceFromMap(res.Data)
//}
//
//func (u *Update) prepareResources(opts *ActionOpts) error {
//	// Extract target version from job resource
//	rel, _, err := extractResources(opts.JobResource.AffectedResources)
//	if err != nil {
//		return err
//	}
//	// Get the job payload struct from the job resource
//	p, err := types.NewPayloadFromString(opts.JobResource.Payload)
//	if err != nil {
//		return err
//	}
//	p.TargetReleaseUUID = rel
//
//	release, err := GetReleaseResource(u.client, p.TargetReleaseUUID)
//	if err != nil {
//		return err
//	}
//	p.TargetResource = release
//	u.jobPayload = p
//	return nil
//}
//
//func (u *Update) Init(optsFn ...ActionOptsFn) error {
//	opts := GetActionOpts(optsFn...)
//	if opts.JobResource == nil || opts.Client == nil {
//		return errors.New("insufficient opts provided to update action")
//	}
//
//	u.client = opts.Client
//
//	if err := u.prepareResources(opts); err != nil {
//		return err
//	}
//
//	if err := u.assertExecutor(); err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (u *Update) GetExecutorName() executors.ExecutorName {
//	return u.executor.GetName()
//}
//
//func (u *Update) assertExecutor() error {
//	ex, err := executors.GetEngineUpdater()
//	if err != nil {
//		return err
//	}
//	log.Infof("action update executor set to: %s", u.GetExecutorName())
//	u.executor = ex
//	return nil
//}
//
//func (u *Update) ExecuteAction() error {
//	//err := u.executor.UpdateEngine(u.client, u.targetRelease)
//	return nil
//}
//
//func (u *Update) GetActionName() ActionName {
//	return UpdateNuvlaEdge
//}
