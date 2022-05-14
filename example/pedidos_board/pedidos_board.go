package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/google/uuid"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Items struct {
	Id       string  `json:"id"`
	Nombre   string  `json:"nombre"`
	Cantidad int     `json:"cantidad"`
	Precio   float64 `json:"precio"`
}

type Pedido struct {
	Id       string  `json:"id"`
	Items    []Items `json:"items"`
	Total    float64
	Fecha    string
	Estado   string `json:"estado"`
	Nombre   string `json:"nombre"`
	ModoPago string `json:"modo_pago"`
	Mesa     string `json:"mesa"`
	Numero   int64
}

var (
	pedidoMutex                    = &sync.Mutex{}
	pedidos     map[string]*Pedido = make(map[string]*Pedido)
	numero      int64              = 0
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger(), middleware.Recover())
	home := liveview.PageControl{
		Title:  "Example2",
		Lang:   "en",
		Path:   "/",
		Router: e,
		HeadCode: `
		<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.2/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-EVSTQN3/azprG1Anm3QDgpJLIm9Nao0Yz1ztcQTwFspd3yD65VohhpuuCOmLASjC" crossorigin="anonymous">
		<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.0.2/dist/js/bootstrap.bundle.min.js" integrity="sha384-MrcW6ZMFYlzcLA8Nl+NtUVF0sA7MsXsP1UyJoMp4YLEuNSfAP+JcXn/tWtIaxVXM" crossorigin="anonymous"></script>
		`,
	}

	e.GET("/pedidos", func(ctx echo.Context) error {
		pedidoMutex.Lock()
		defer pedidoMutex.Unlock()

		return ctx.JSON(http.StatusOK, pedidos)
	})
	e.GET("/pedidos/:id", func(ctx echo.Context) error {
		pedidoMutex.Lock()
		defer pedidoMutex.Unlock()
		id := ctx.Param("id")
		if pedido, ok := pedidos[id]; ok {
			return ctx.JSON(http.StatusOK, pedido)
		}
		return ctx.JSON(http.StatusNotFound, "Pedido no encontrado")
	})
	e.POST("/pedidos", func(ctx echo.Context) error {
		defer liveview.SendToAllLayouts("EVENT_UPDATE_PEDIDOS")
		pedidoMutex.Lock()
		defer pedidoMutex.Unlock()
		var pedido Pedido
		if err := json.NewDecoder(ctx.Request().Body).Decode(&pedido); err != nil {
			ctx.JSON(400, err)
			return err
		}
		pedido.Fecha = time.Now().Format("2006-01-02 15:04:05")
		pedido.Estado = "Nuevo"
		pedido.Id = uuid.NewString()
		numero++
		pedido.Numero = numero
		pedidos[pedido.Id] = &pedido
		return ctx.JSON(http.StatusOK, pedido)
	})

	e.PUT("/pedidos/:id", func(ctx echo.Context) error {
		defer liveview.SendToAllLayouts("EVENT_UPDATE_PEDIDOS")
		pedidoMutex.Lock()
		defer pedidoMutex.Unlock()
		id := ctx.Param("id")
		status := ctx.FormValue("estado")
		if _, ok := pedidos[id]; !ok {
			return ctx.JSON(http.StatusNotFound, "pedido no encontrado")
		}

		if status == "" {
			return ctx.JSON(http.StatusBadRequest, "estado es requerido")
		}

		if strings.ToLower(status) == "procesando" && pedidos[id].Estado == "Nuevo" {
			pedidos[id].Estado = "Procesando"
		}
		if strings.ToLower(status) == "listo" && pedidos[id].Estado == "Procesando" {
			pedidos[id].Estado = "Listo"
		}

		if strings.ToLower(status) == "cancelado" {
			delete(pedidos, id)
			return ctx.JSON(http.StatusOK, "Pedido cancelado")
		}
		return ctx.JSON(http.StatusOK, pedidos[id])
	})
	e.DELETE("/pedidos/{id}", func(ctx echo.Context) error {
		defer liveview.SendToAllLayouts("EVENT_UPDATE_PEDIDOS")
		pedidoMutex.Lock()
		defer pedidoMutex.Unlock()
		id := ctx.Param("id")
		delete(pedidos, id)
		return ctx.JSON(http.StatusNoContent, "")
	})

	home.Register(func() liveview.LiveDriver {
		document := liveview.NewLayout(`

		{{mount "button_send"}}
		<hr/>
		<ul class="nav nav-tabs" id="myTab" role="tablist">
		<li class="nav-item" role="presentation">
		  <button class="nav-link active" id="nuevos-tab" data-bs-toggle="tab" data-bs-target="#nuevos" type="button" role="tab" aria-controls="nuevos" aria-selected="true">Nuevos</button>
		</li>
		<li class="nav-item" role="presentation">
		  <button class="nav-link" id="procesando-tab" data-bs-toggle="tab" data-bs-target="#procesando" type="button" role="tab" aria-controls="procesando" aria-selected="false">Procesando</button>
		</li>
		<li class="nav-item" role="presentation">
		  <button class="nav-link" id="terminados-tab" data-bs-toggle="tab" data-bs-target="#terminados" type="button" role="tab" aria-controls="terminados" aria-selected="false">Terminados</button>
		</li>
		<li class="nav-item" role="presentation">
		<button class="nav-link" id="cancelados-tab" data-bs-toggle="tab" data-bs-target="#cancelados" type="button" role="tab" aria-controls="cancelados" aria-selected="false">Cancelados</button>
	  </li>
	  </ul>

	  <div class="tab-content" id="myTabContent">
		<div class="tab-pane fade show active" id="nuevos" role="tabpanel" aria-labelledby="nuevos-tab"><div id="div_pedidos_nuevos"></div></div>
		<div class="tab-pane fade" id="procesando" role="tabpanel" aria-labelledby="procesando-tab"><div id="div_pedidos_procesando"></div></div>
		<div class="tab-pane fade" id="terminados" role="tabpanel" aria-labelledby="terminados-tab"><div id="div_pedidos_terminados"></div></div>
		<div class="tab-pane fade" id="cancelados" role="tabpanel" aria-labelledby="cancelados-tab"><div id="div_pedidos_cancelados"></div></div>
	  </div>

			
			<hr/>
			
			<hr/>
			<div id="div_status"></div>`)
		liveview.New("button_send", &components.Button{Caption: "Actualizar"}).
			SetClick(func(this *components.Button, data interface{}) {
				liveview.SendToAllLayouts("EVENT_UPDATE_PEDIDOS")
			})
		document.Component.SetHandlerEventIn(func(data interface{}) {
			pedidoMutex.Lock()
			defer pedidoMutex.Unlock()
			msg := data.(string)
			if msg == "EVENT_UPDATE_PEDIDOS" || msg == "FIRST_TIME" {
				divGeneral := document.GetDriverById("div_pedidos_nuevos")
				html := `<div class="d-flex flex-row  mb-3">`
				i := 0
				ppedidos := make([]*Pedido, len(pedidos))
				for _, pedido := range pedidos {
					ppedidos[i] = pedido
					i++
				}
				sort.Slice(ppedidos, func(i, j int) bool {
					return ppedidos[i].Numero < ppedidos[j].Numero
				})

				i = 0
				for _, pedido := range ppedidos {
					i++
					if i%5 == 0 {
						i = 1
						html += `</div>`
						html += `<div class="d-flex flex-row  mb-3">`
					}
					htmlItems := ""
					total := 0.0
					for _, item := range pedido.Items {
						htmlItems += fmt.Sprintf(`<tr><td>%s</td><td>%d</td><td>$%.2f</td></tr>`, item.Nombre, item.Cantidad, item.Precio)
						total += item.Precio
					}
					pedido.Total = total
					html += fmt.Sprintf(`
						<div class="card mt-2 ms-2  " style="width: 18rem;">
							<div class="card-header">
								<h5 class="card-title">Mesa:%s Numero:%d</h5>
							</div>
							<div class="card-body">
								<p class="card-text">
									<table class="table table-striped">
									<thead><tr><th>Producto</th><th>Cantidad</th><th>Precio</th></tr></thead>
									<tbody>%s</tbody>
									</table>
									<b>Estado:</b> %s<br/>
									<b>Total:</b> $%.2f
								</p>
							</div>
						</div>`, pedido.Mesa, pedido.Numero, htmlItems, pedido.Estado, pedido.Total)
				}
				html += `</div>`
				divGeneral.FillValue(html)
			}
		})

		/*
			document.Component.SetHandlerFirstTime(func() {
				liveview.SendToAllLayouts("EVENT_UPDATE_PEDIDOS")
			})
		//*/

		document.Component.SetHandlerEventTime(time.Second*5, func() {
			spanGlobalStatus := document.GetDriverById("div_status")
			spanGlobalStatus.FillValue("online")
		})

		document.Component.SetHandlerEventDestroy(func(id string) {
			liveview.SendToAllLayouts("EVENT_UPDATE_PEDIDOS")
		})
		return document
	})
	e.Logger.Fatal(e.Start(":1323"))
}
