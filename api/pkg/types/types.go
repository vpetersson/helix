package types

import (
	"context"
	"time"
)

type BalanceTransferData struct {
	JobID           string `json:"job_id"`
	StripePaymentID string `json:"stripe_payment_id"`
}

type BalanceTransfer struct {
	ID          string              `json:"id"`
	Created     time.Time           `json:"created"`
	Owner       string              `json:"owner"`
	OwnerType   OwnerType           `json:"owner_type"`
	PaymentType PaymentType         `json:"payment_type"`
	Amount      int                 `json:"amount"`
	Data        BalanceTransferData `json:"data"`
}

type Module struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Cost     int    `json:"cost"`
	Template string `json:"template"`
}

type Interaction struct {
	ID      string      `json:"id"`
	Created time.Time   `json:"created"`
	Creator CreatorType `json:"creator"` // e.g. User
	// the ID of the runner that processed this interaction
	Runner   string            `json:"runner"`   // e.g. 0
	Message  string            `json:"message"`  // e.g. Prove pythagoras
	Progress int               `json:"progress"` // e.g. 0-100
	Files    []string          `json:"files"`    // list of filepath paths
	Finished bool              `json:"finished"` // if true, the message has finished being written to, and is ready for a response (e.g. from the other participant)
	Metadata map[string]string `json:"metadata"` // different modes and models can put values here - for example, the image fine tuning will keep labels here to display in the frontend
	Error    string            `json:"error"`
	// we hoist this from files so a single interaction knows that it "Created a finetune file"
	FinetuneFile string `json:"finetune_file"`
}

type Session struct {
	ID string `json:"id"`
	// name that goes in the UI - ideally autogenerated by AI but for now can be
	// named manually
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	// e.g. inference, finetune
	Mode SessionMode `json:"mode"`
	// e.g. text, image
	Type SessionType `json:"type"`
	// huggingface model name e.g. mistralai/Mistral-7B-Instruct-v0.1 or
	// stabilityai/stable-diffusion-xl-base-1.0
	ModelName ModelName `json:"model_name"`
	// if type == finetune, we record a filestore path to e.g. lora file here
	// currently the only place you can do inference on a finetune is within the
	// session where the finetune was generated
	FinetuneFile string `json:"finetune_file"`
	// for now we just whack the entire history of the interaction in here, json
	// style
	Interactions []Interaction `json:"interactions"`
	// uuid of owner entity
	Owner string `json:"owner"`
	// e.g. user, system, org
	OwnerType OwnerType `json:"owner_type"`
}

type SessionFilterModel struct {
	Mode         SessionMode `json:"mode"`
	ModelName    ModelName   `json:"model_name"`
	FinetuneFile string      `json:"finetune_file"`
}

type SessionFilter struct {
	// e.g. inference, finetune
	Mode SessionMode `json:"mode"`
	// e.g. text, image
	Type SessionType `json:"type"`
	// huggingface model name e.g. mistralai/Mistral-7B-Instruct-v0.1 or
	// stabilityai/stable-diffusion-xl-base-1.0
	ModelName ModelName `json:"model_name"`
	// the filestore path to the file being used for finetuning
	FinetuneFile string `json:"finetune_file"`
	// this means "only give me sessions that will fit in this much ram"
	Memory uint64 `json:"memory"`

	// the list of model name / mode combos that we should skip over
	// normally used by runners that are running multiple types in parallel
	// who don't want another version of what they are already running
	Reject []SessionFilterModel `json:"reject"`
}

type ApiKey struct {
	Owner     string    `json:"owner"`
	OwnerType OwnerType `json:"owner_type"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
}

// passed between the api server and the controller
type RequestContext struct {
	Ctx       context.Context
	Owner     string
	OwnerType OwnerType
}

type UserStatus struct {
	User    string `json:"user"`
	Credits int    `json:"credits"`
}

type WebsocketEvent struct {
	Type    WebsocketEventType `json:"type"`
	Session *Session           `json:"session"`
}

// the context of a long running python process
// on a runner - this will be used to inject the env
// into the cmd returned by the model instance.GetCommand() function
type RunnerProcessConfig struct {
	// the id of the model instance
	InstanceID string `json:"instance_id"`
	// the URL to ask for more tasks
	// this will pop the task from the queue
	TaskURL string `json:"task_url"`
	// the URL to ask for what the session is (e.g. to know what finetune_file to load)
	// this is readonly and will not pop the session(task) from the queue
	SessionURL string `json:"session_url"`
	// the URL to send responses to
	ResponseURL string `json:"response_url"`
}

// a session will run "tasks" on runners
// task's job is to take the most recent user interaction
// and add a response to it in the form of a system interaction
// the api controller will have already appended the system interaction
// to the very end of the Session.Interactions list
// our job is to fill in the Message and/or Files field of that interaction
type WorkerTask struct {
	SessionID string `json:"session_id"`
	// the string that we are calling the prompt that we will feed into the model
	Prompt string `json:"prompt"`

	// this is the Lora type file that a fine tune session produced
	// it is used for inference against a fine tuned model
	FinetuneFile string `json:"finetune_file"`

	// this is the directory that contains the files used for fine tuning
	// i.e. it's the user files that will be the input to a finetune session
	FinetuneInputDir string `json:"finetune_input_dir"`
}

type WorkerTaskResponse struct {
	// the python code must submit these fields back to the runner api
	Type      WorkerTaskResponseType `json:"type"`
	SessionID string                 `json:"session_id"`
	// this is filled in by the runner on the way back to the api
	InteractionID string `json:"interaction_id"`
	// which fields the python code decides to fill in here depends
	// on what the type of model it is
	Message  string   `json:"message"`  // e.g. Prove pythagoras
	Progress int      `json:"progress"` // e.g. 0-100
	Files    []string `json:"files"`    // list of filepath paths
	Error    string   `json:"error"`
}

type ServerConfig struct {
	// used to prepend onto raw filestore paths to download files
	// the filestore path will have the user info in it - i.e.
	// it's a low level filestore path
	// if we are using an object storage thing - then this URL
	// can be the prefix to the bucket
	FilestorePrefix string `json:"filestore_prefix"`
}
