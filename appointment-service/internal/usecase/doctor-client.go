package usecase

type Doctor_client interface {
	Doctor_exists(doctor_ID string) (bool, error)
}
