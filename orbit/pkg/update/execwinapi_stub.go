//go:build !windows

package update

func RunWindowsMDMEnrollment(args WindowsMDMEnrollmentArgs) error {
	return nil
}

func RunWindowsMDMUnenrollment(args WindowsMDMEnrollmentArgs) error {
	return nil
}
