package api

import (
	"net/http"

	"github.com/astronomerio/asds/marathon"
	"github.com/astronomerio/asds/mesos"
	"github.com/astronomerio/asds/pkg"
	"github.com/astronomerio/asds/provisioner"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// Handler client
type Handler struct {
	log      *logrus.Entry
	marathon *marathon.Client
	mesos    *mesos.Client
	pi       *pkg.PrometheusInstrumentation
}

// NewHandler returns a new Handler Client
func NewHandler(log *logrus.Logger, marathon *marathon.Client, mesos *mesos.Client, pi *pkg.PrometheusInstrumentation) *Handler {
	return &Handler{
		log:      log.WithFields(logrus.Fields{"package": "api"}),
		marathon: marathon,
		mesos:    mesos,
		pi:       pi,
	}
}

// getOrg returns information
func (h *Handler) getOrg(c *gin.Context) {
	logger := h.log.WithFields(logrus.Fields{"function": "getOrg"})
	logger.Debug("Entered getOrg")

	orgID := c.Param("id")
	logger.WithField("id", orgID).Debug("Read param")

	var prov provisioner.Provisioner
	prov = h.marathon // We can change provisioners in the future
	resp, err := prov.GetProvision(&provisioner.ProvisionRequest{OrgID: orgID})
	if err != nil {
		h.log.Error(err)
		c.JSON(http.StatusInternalServerError, resp)
	} else if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Organization " + orgID + " not found"})
	} else {
		c.JSON(http.StatusOK, resp)
	}

	logger.Debug("Exiting getOrg")
}

// postOrg creates a new orginization deployment.  It will coll the Provisioner.Provision method.
func (h *Handler) postOrg(c *gin.Context) {
	logger := h.log.WithFields(logrus.Fields{"function": "postOrg"})
	logger.Debug("Entered postOrg")

	provRqst := &provisioner.ProvisionRequest{}
	orgIDParam := c.Param("id")
	logger.WithField("orgIDParam", orgIDParam).Debug("got request")
	if len(orgIDParam) == 0 {
		if err := c.BindJSON(provRqst); err != nil {
			logger.Error(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	} else {
		provRqst.OrgID = orgIDParam
	}

	var prov provisioner.Provisioner
	prov = h.marathon // We can change provisioners in the future
	resp, err := prov.Provision(provRqst)

	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	h.pi.ProProvisionCount.Inc()
	c.JSON(http.StatusOK, resp)

	logger.Debug("Exiting postOrg")
}

// postOrgReprovision will reprovision the given orginization
func (h *Handler) postOrgReprovision(c *gin.Context) {
	logger := h.log.WithFields(logrus.Fields{"function": "postOrgReprovision"})
	logger.Debug("Entered postOrg")

	provRqst := &provisioner.ProvisionRequest{}
	orgIDParam := c.Param("id")
	logger.WithField("orgIDParam", orgIDParam).Debug("got request")
	if len(orgIDParam) == 0 {
		if err := c.BindJSON(provRqst); err != nil {
			logger.Error(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	} else {
		provRqst.OrgID = orgIDParam
	}

	var prov provisioner.Provisioner
	prov = h.marathon // We can change provisioners in the future
	resp, err := prov.ReProvision(provRqst)

	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	c.JSON(http.StatusOK, resp)

	logger.Debug("Exiting postOrgReprovision")
}

// deleteOrg will delete the given orginization and all its provisions
func (h *Handler) deleteOrg(c *gin.Context) {
	logger := h.log.WithFields(logrus.Fields{"function": "deleteOrd"})
	logger.Debug("Entered deleteOrg")

	provRqst := &provisioner.ProvisionRequest{}
	orgIDParam := c.Param("id")
	logger.WithField("orgIDParam", orgIDParam).Debug("got request")
	if len(orgIDParam) == 0 {
		if err := c.BindJSON(provRqst); err != nil {
			logger.Error(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	} else {
		provRqst.OrgID = orgIDParam
	}

	logger.WithField("orgID", provRqst).Debug("Got OrgID")

	var prov provisioner.Provisioner
	prov = h.marathon // We can change provisioners in the future
	resp, err := prov.RemoveProvision(provRqst)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	h.pi.ProDeprovisionCount.Inc()
	c.JSON(http.StatusOK, resp)

	logger.Debug("Exiting postOrg")
}

func (h *Handler) deployDag(c *gin.Context) {
	logger := h.log.WithField("function", "deployDag")
	logger.Debug("Entered deployDag")

	provRqst := &provisioner.ProvisionRequest{}
	if err := c.BindJSON(provRqst); err != nil {
		logger.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var prov provisioner.Provisioner
	prov = h.marathon
	resp, err := prov.DeployDag(provRqst)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	h.pi.ProDeploymentCount.With(prometheus.Labels{"orgid": provRqst.OrgID}).Inc()
	c.JSON(http.StatusOK, resp)

}

// getApps is used for testing Marathon, returns all apps in Marathon.
func (h *Handler) getApps(c *gin.Context) {
	logger := h.log.WithField("function", "getApps")
	logger.Debug("Entered getApps")
	apps, err := h.marathon.GetAllApplications()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.log.Error(err)
	} else {
		c.JSON(http.StatusOK, apps)
	}
}

func (h *Handler) getProvisions(c *gin.Context) {
	logger := h.log.WithField("function", "getProvisions")
	logger.Debug("Entered getProvisions")

	var prov provisioner.Provisioner
	prov = h.marathon
	resp, err := prov.ListProvisions()
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) getAvailableResources(c *gin.Context) {
	logger := h.log.WithField("function", "getAvailableResources")
	logger.Debug("Entered getAvailableResources")

	resp, err := h.mesos.GetAvailableResources()
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) getMetrics(c *gin.Context) {
	// Get Mesos Metrics
	h.mesos.Metrics()
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
