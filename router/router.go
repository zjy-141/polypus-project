package router

import (
	"polypus-project/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.Use(middleware.Error)
	r.Use(middleware.GinLogger(), middleware.GinRecovery(true))
	apiRouter := r.Group("/api")
	{
		// example
		// begin
		userRouter := apiRouter.Group("/user")
		{
			userRouter.POST("/login", ctr.User.Login)
			userRouter := userRouter.Use(middleware.CheckRole(0))
			userRouter.GET("/status", ctr.User.Status)
			userRouter.DELETE("/logout", ctr.User.Logout)
			userRouter.PUT("/update", ctr.User.Update)
		}
		adminRouter := apiRouter.Group("/admin").Use(middleware.CheckRole(3))
		{
			adminRouter.POST("/doctor/register", ctr.Admin.DoctorRegister)
			// adminRouter.PUT("/doctor/upgrade/:doctorid", ctr.Admin.DoctorUpgrade)
			adminRouter.GET("/doctor/show", ctr.Admin.DoctorShow)
			adminRouter.PUT("/doctor/reset/:doctorid", ctr.Admin.DoctorReset)
			adminRouter.DELETE("/doctor/delete/:doctorid", ctr.Admin.DoctorDelete)
		}
		doctorRouter := apiRouter.Group("/doctor").Use(middleware.CheckRole(2))
		{
			doctorRouter.POST("/patient/register", ctr.Form.PatientRegister)
			doctorRouter.POST("/input", ctr.Form.FormInput)
			doctorRouter.POST("/save", ctr.Form.FormSave)
			doctorRouter.GET("/patient/history", ctr.Form.PatientHistory)
			doctorRouter.PUT("/update", ctr.Form.FormUpdate)
			doctorRouter.DELETE("/delete/:formid", ctr.Form.FormDelete)
		}
		backstageRouter := apiRouter.Group("/backstage").Use(middleware.CheckRole(2))
		{
			backstageRouter.GET("/get", ctr.Form.GetAll)
		}
		// end
	}
}
