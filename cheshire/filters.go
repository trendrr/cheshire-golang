package cheshire

import (
	"log"
	"net/http"
)



// Hooks to hook into before and after the controller execution.
type ControllerFilter interface {
	//This is called before the Controller is called. 
	//returning false will stop the execution
	Before(*Txn) bool
}

// Additional hooks if you need more granularity into the lifecycle
type FilterAdvanced interface {
	ControllerFilter

	//Called immediately before the response is written.
	BeforeWrite(*Response, *Txn)

	//This is called after the controller is called.
	//The response has already been sent 
	AfterWrite(*Response, *Txn)
}


// A generic cache.
type Cache interface {
	Set(key string, value []byte, expireSeconds int)

	// Sets the value if and only if there is no value associated with this key
	SetIfAbsent(key string, value []byte, expireSeconds int) bool

	// Deletes the value at the requested key
	Delete(key string)

	// Gets the value at the requested key
	Get(key string) ([]byte, bool)

	// Increment the key by val (val is allowed to be negative)
	// in most implementation expireSeconds will be from the first increment, but users should not count on that.
	// if no value is a present it should be added.  
	// If a value is present which is not a number an error should be returned.
	Inc(key string, val int64, expireSeconds int) (int64, error)
}


type Session struct {
	//The cache provider
	cache Cache

	//Max age for the session
	sessionAgeSeconds int
}

func NewSession(cache Cache, sessionMaxSeconds int) *Session {
	return &Session{
		cache:             cache,
		sessionAgeSeconds: sessionMaxSeconds,
	}
}

//This is called before the Controller is called. 
//returning false will stop the execution
func (this *Session) Before(txn *Txn) bool {
	if txn.Type() != "html" {
		//skip
		return true
	}
	// log.Println("SESSION!")
	httpWriter, err := ToHttpWriter(txn)
	if err != nil {
		log.Printf("ERROR in session.before %s", err)
		return true //should we continue with the request?
	}

	cookie, err := httpWriter.HttpRequest.Cookie("session_id")

	if err != nil {
		//create new session id
		txn.Session.Put("session_id", SessionId())
		// log.Println("Created session!")
	} else {
		//load the session. 
		sessionId := cookie.Value
		// log.Printf("Found session cookie! %s", sessionId)
		bytes, ok := this.cache.Get(sessionId)
		if !ok {
			//create a new session, since the old one is gone
			sessionId = SessionId()
			// log.Printf("Old session expired, setting new one (%s)", sessionId)
		} else {
			err = txn.Session.UnmarshalJSON(bytes)
			if err != nil {
				log.Printf("Error unmarshaling json (%s) -> (%s)", bytes, err)
			}
		}
		txn.Session.Put("session_id", sessionId)
	}
	return true
}

func (this *Session) BeforeHtmlWrite(txn *Txn, writer http.ResponseWriter) bool {

	sessionId, ok := txn.Session.GetString("session_id")
	if !ok {
		log.Println("Error! No Sessionid in txn.  wtf?")
		return true
	}
	if txn.Session.MustBool("delete_session", false) {
		log.Println("Deleting session")
		cookie := &http.Cookie{
			Name:   "session_id",
			Value:  sessionId,
			MaxAge: 0,
		}
		http.SetCookie(writer, cookie)
		this.cache.Delete(sessionId)
		return true
	}

	//session will always have session_id param
	if len(txn.Session.Map) > 1 {
        //We set an internal flag in the session, so 
        //if keys are removed now we always save the session. 
        txn.Session.Put("_persisted", true)

		//only write the session if there is something in it
		cookie := &http.Cookie{
			Name:   "session_id",
			Value:  sessionId,
			MaxAge: this.sessionAgeSeconds,
		}
		http.SetCookie(writer, cookie)
		bytes, err := txn.Session.MarshalJSON()
		if err != nil {
			log.Println(err)
		}
		this.cache.Set(sessionId, bytes, this.sessionAgeSeconds)
	}
	return true
}

// returns a unique session id
func SessionId() string {
	id := RandString(32)
	//TODO; this isnt 
	return id
}
