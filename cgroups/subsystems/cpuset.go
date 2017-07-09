package subsystems

type CpusetSubsystem struct {
}

func (s *CpusetSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	return nil
}

func (s *CpusetSubsystem) Apply(cgroupPath string, pid int) error {
	return nil
}

func (s *CpusetSubsystem) Remove(cgroupPath string) error {
	return nil
}

func (s *CpusetSubsystem) Name() string {
	return nil
}
