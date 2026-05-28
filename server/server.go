package server

import (
	httpin_integration "github.com/ggicci/httpin/integration"
)

func init() {
	httpin_integration.UseHttpPathVariable("path")
}
