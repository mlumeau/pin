package export

import "net/http"

// Register wires identity export routes (json/xml/txt/vcf).
func Register(register func(pattern string, handler http.Handler), source Source) {
	handler := NewHandler(source)

	// Owner identity exports only.
	register("/.well-known/identity.json", http.HandlerFunc(handler.IdentityJSON))
	register("/.well-known/identity.xml", http.HandlerFunc(handler.IdentityXML))
	register("/.well-known/identity.txt", http.HandlerFunc(handler.IdentityTXT))
	register("/.well-known/identity.vcf", http.HandlerFunc(handler.IdentityVCF))
}
