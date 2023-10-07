package rockset

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func addOptionIfChanged[O any](d *schema.ResourceData, key string, options *[]O, fn func(any) O) {
	if d.HasChange(key) {
		*options = append(*options, fn(d.Get(key)))
	}
}

func setValue[T any](d *schema.ResourceData, key string, fn func() (T, bool)) error {
	if v, ok := fn(); ok {
		if err := d.Set(key, v); err != nil {
			return err
		}
	}

	return nil
}
