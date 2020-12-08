package patient

func (c AdmitPatient) IsCommand()     {}
func (c TransferPatient) IsCommand()  {}
func (c DischargePatient) IsCommand() {}

// PatientAdmitted event.
type AdmitPatient struct {
	ID   string     `json:"id"`
	Name Name       `json:"name"`
	Ward WardNumber `json:"ward"`
	Age  Age        `json:"age"`
}

// PatientTransferred event.
type TransferPatient struct {
	ID            string     `json:"id"`
	NewWardNumber WardNumber `json:"new_ward"`
}

// PatientDischarged event.
type DischargePatient struct {
	ID string `json:"id"`
}
