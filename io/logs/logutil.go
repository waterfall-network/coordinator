// Package logs creates a Multi writer instance that
// write all logs that are written to stdout.
package logs

import (
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
)

func addLogWriter(w io.Writer) {
	mw := io.MultiWriter(logrus.StandardLogger().Out, w)
	logrus.SetOutput(mw)
}

// ConfigurePersistentLogging adds a log-to-file writer. File content is identical to stdout.
func ConfigurePersistentLogging(logFileName string) error {
	logrus.WithField("logFileName", logFileName).Info("Logs will be made persistent")
	f, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, params.BeaconIoConfig().ReadWritePermissions) // #nosec G304
	if err != nil {
		return err
	}

	addLogWriter(f)

	logrus.Info("File logging initialized")
	return nil
}

// MaskCredentialsLogging masks the url credentials before logging for security purpose
// [scheme:][//[userinfo@]host][/]path[?query][#fragment] -->  [scheme:][//[***]host][/***][#***]
// if the format is not matched nothing is done, string is returned as is.
func MaskCredentialsLogging(currURL string) string {
	// error if the input is not a URL
	MaskedURL := currURL
	u, err := url.Parse(currURL)
	if err != nil {
		return currURL // Not a URL, nothing to do
	}
	// Mask the userinfo and the URI (path?query or opaque?query ) and fragment, leave the scheme and host(host/port)  untouched
	if u.User != nil {
		MaskedURL = strings.Replace(MaskedURL, u.User.String(), "***", 1)
	}
	if len(u.RequestURI()) > 1 { // Ignore the '/'
		MaskedURL = strings.Replace(MaskedURL, u.RequestURI(), "/***", 1)
	}
	if len(u.Fragment) > 0 {
		MaskedURL = strings.Replace(MaskedURL, u.RawFragment, "***", 1)
	}
	return MaskedURL
}
