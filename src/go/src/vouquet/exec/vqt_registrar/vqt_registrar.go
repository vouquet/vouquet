package main

import (
	"vouquet/soil"
)

func main() {
	r, err := soil.OpenRegister("./sqlconfigpath")
	if err != nil {
		return err
	}
	defer r.Close()

	for _, s := range soil.SOIL_ALL {
		t, err := soil.NewThemograpy(s)
		if err != nil {
			return err
		}

		go func () {
			for {
				if err := r.Record(t); err != nil {
					return
				}
			}
		}()
	}
}
