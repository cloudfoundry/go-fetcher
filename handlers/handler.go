package handlers

import (
  "net/http"
	"strings"
	"fmt"
)

type handler struct {
	domain  *string
	orgList []string
}

func NewHandler(domain *string, orgList []string) *handler {
  return &handler{
		domain: domain,
		orgList: orgList,
	}
}

func (h *handler) GetMeta(writer http.ResponseWriter, request *http.Request) {
		repoName := strings.Split(request.URL.Path, "/")[1]
		fmt.Fprintf(writer, "<meta name=\"go-import\" content=\"%s git %s\">",
			*h.domain + "/" + repoName,
			h.orgList[0] + repoName,
		)
}
