package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/go-chi/chi/v5"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type createContainerInput struct {
	Image string `json:"containerImage"`
	Tag   string `json:"imageTag"`
}

func CreateRouter(cli *client.Client) http.Handler {
	router := chi.NewRouter()

	router.Post("/create", func(w http.ResponseWriter, r *http.Request) {
		var input createContainerInput
		if err := decode(r, &input); err != nil {
			Respond(w, http.StatusBadRequest, "Invalid input")
			return
		}

		images, err := cli.ImageList(r.Context(), image.ListOptions{All: false})
		if err != nil {
			Respond(w, http.StatusInternalServerError, err.Error())
			return
		}

		var imageExists bool
		var img = input.Image + ":" + input.Tag

	outer:
		for _, image := range images {
			for _, tag := range image.RepoTags {
				if tag == img {
					imageExists = true
					break outer
				}
			}
		}

		if !imageExists {
			reader, err := cli.ImagePull(r.Context(), img, image.PullOptions{})
			if err != nil {
				Respond(w, http.StatusInternalServerError, err.Error())
				return
			}

			defer reader.Close()
			io.Copy(os.Stdout, reader)
		}

		println("Creating container...")
		resp, err := cli.ContainerCreate(
			r.Context(),
			&container.Config{Tty: false, Image: img},
			&container.HostConfig{AutoRemove: true},
			&network.NetworkingConfig{},
			&v1.Platform{},
			"",
		)
		if err != nil {
			Respond(w, http.StatusInternalServerError, err.Error())
			return
		}

		println("Starting container...")
		if err := cli.ContainerStart(r.Context(), resp.ID, container.StartOptions{}); err != nil {
			Respond(w, http.StatusInternalServerError, err.Error())
			return
		}

		info, err := cli.ContainerInspect(r.Context(), resp.ID)
		if err != nil {
			Respond(w, http.StatusInternalServerError, err.Error())
			return
		}

		Respond(w, http.StatusCreated, map[string]string{"url": info.Name[1:] + ".localhost:8000"})
	})

	return router
}

func decode(r *http.Request, val any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(val)
}

func Respond(w http.ResponseWriter, statusCode int, data any) error {
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(data)
}
