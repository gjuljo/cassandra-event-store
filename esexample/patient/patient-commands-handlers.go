package patient

import "my/esexample/store"

type PatientCommandHandler struct {
	store *patientEventStore
}

func NewPatientCommandHandler(store *patientEventStore) *PatientCommandHandler {
	return &PatientCommandHandler{store: store}
}

func (h *PatientCommandHandler) HandleAdmitPatient(c *AdmitPatient) error {
	p := New(c.ID, c.Name, c.Age, c.Ward)
	return h.store.Update(p)
}

func (h *PatientCommandHandler) HandleTransferPatient(c *TransferPatient) error {
	p, err := h.store.Find(store.EventID(c.ID))

	if err != nil {
		return err
	}

	err = p.Transfer(c.NewWardNumber)

	if err != nil {
		return err
	}

	return h.store.Update(p)
}

func (h *PatientCommandHandler) HandleDischargePatient(c *DischargePatient) error {
	p, err := h.store.Find(store.EventID(c.ID))

	if err != nil {
		return err
	}

	err = p.Discharge()

	if err != nil {
		return err
	}

	return h.store.Update(p)
}
