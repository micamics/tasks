package main

import (
	"fmt"
	"net/http"
	"strconv"

	sdkapi "sdk/api"
)

func main() {}

func Init(api sdkapi.IPluginApi) {
	// Your plugin code here
	httpAPI := api.Http()
	pluginConfigAPI := api.Config().Plugin()
	adminRouter := httpAPI.HttpRouter().AdminRouter()

	// Define the settings form
	settingsForm := sdkapi.HttpForm{
		Name:          "settings",
		CallbackRoute: "settings.save",
		SubmitLabel:   "Submit",
		Sections: []sdkapi.FormSection{
			{
				Name: "General Configuration",
				Fields: []sdkapi.IFormField{
					sdkapi.FormTextField{
						Name:  "Banner Text",
						Label: "Banner Text",
						ValueFn: func() string {
							b, err := pluginConfigAPI.Read("banner_text")
							if err != nil {
								return "This is the default banner text!"
							}
							return string(b)
						},
					},
					sdkapi.FormIntegerField{
						Name:  "Integer Field",
						Label: "Integer Field",
						ValueFn: func() int64 {
							defaultVal := int64(3)
							val, err := pluginConfigAPI.Read("integer_field")
							if err != nil {
								return defaultVal
							}

							num, err := strconv.Atoi(string(val))
							if err == nil {
								return int64(num)
							}

							return defaultVal
						},
					},
					sdkapi.FormBooleanField{
						Name:  "Boolean Field",
						Label: "Boolean Field",
						ValueFn: func() bool {
							defaultVal := true
							val, err := pluginConfigAPI.Read("boolean_field")
							if err != nil {
								return defaultVal
							}

							boolVal, err := strconv.ParseBool(string(val))
							if err == nil {
								return boolVal
							}

							return defaultVal
						},
					},
				},
			},
		},
	}

	// register the settings form
	if err := httpAPI.Forms().RegisterForms(settingsForm); err != nil {
		api.Logger().Error("Failed to register settings form: %s", err)
		return
	}

	// Add a new route group to the admin router
	adminRouter.Group("/settings", func(subrouter sdkapi.IHttpRouterInstance) {

		// Show the settings form
		subrouter.Get("/form", func(w http.ResponseWriter, r *http.Request) {
			// Retrieve the registered form
			form, ok := httpAPI.Forms().GetForm("settings")
			if !ok {
				httpAPI.HttpResponse().Error(w, r, fmt.Errorf("Form not found"), http.StatusInternalServerError)
				return
			}

			// Render the form template
			htmlForm := form.GetTemplate(r)
			httpAPI.HttpResponse().AdminView(w, r, sdkapi.ViewPage{PageContent: htmlForm})
		}).Name("settings.form")

		// Save the settings
		subrouter.Post("/save", func(w http.ResponseWriter, r *http.Request) {
			// Retrieve the registered settings form
			form, ok := httpAPI.Forms().GetForm("settings")
			if !ok {
				httpAPI.HttpResponse().Error(w, r, fmt.Errorf("Form not found"), http.StatusInternalServerError)
				return
			}

			if err := form.ParseForm(r); err != nil {
				httpAPI.HttpResponse().Error(w, r, err, http.StatusInternalServerError)
				return
			}

			bannerText, err := form.GetStringValue("General Configuration", "Banner Text")
			if err != nil {
				httpAPI.HttpResponse().Error(w, r, err, http.StatusInternalServerError)
				return
			}

			intVal, err := form.GetIntValue("General Configuration", "Integer Field")
			if err != nil {
				httpAPI.HttpResponse().Error(w, r, err, http.StatusInternalServerError)
				return
			}

			boolVal, err := form.GetBoolValue("General Configuration", "Boolean Field")
			if err != nil {
				httpAPI.HttpResponse().Error(w, r, err, http.StatusInternalServerError)
				return
			}

			// Write the new value to the plugin configuration and send a success message
			pluginConfigAPI.Write("banner_text", []byte(bannerText))
			pluginConfigAPI.Write("integer_field", []byte(fmt.Sprint(intVal)))
			pluginConfigAPI.Write("boolean_field", []byte(fmt.Sprint(boolVal)))
			httpAPI.HttpResponse().FlashMsg(w, r, "Settings saved successfully", sdkapi.FlashMsgSuccess)
			httpAPI.HttpResponse().Redirect(w, r, "settings.form")

		}).Name("settings.save")
	})

	// Register navigation menu items
	httpAPI.Navs().AdminNavsFactory(func(r *http.Request) []sdkapi.AdminNavItemOpt {
		return []sdkapi.AdminNavItemOpt{
			{
				Category:  sdkapi.NavCategorySystem,
				Label:     "Sample Plugin",
				RouteName: "settings.form",
			},
		}
	})
}
