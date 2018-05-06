package doctor

func (d *Doctor) extractVersion(info *ClusterInfo) error {
	v, err := d.kc.Discovery().ServerVersion()
	if err != nil {
		return err
	}
	info.Version = *v
	return err
}
