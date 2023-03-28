package storyblok

// https://www.storyblok.com/docs/api/content-delivery
var statusCodes = map[int]string{
	200: "OK Everything worked as expected.",
	400: "Bad Request Wrong format was sent (eg. XML instead of JSON).",
	401: "Unauthorized No valid API key provided.",
	404: "Not Found	The requested resource doesn't exist (perhaps due to not yet published content entries).",
	422: "Unprocessable Entity The request was unacceptable, often due to missing a required parameter.",
	429: "Too Many Requests	Too many requests hit the API too quickly. We recommend an exponential backoff of your requests.",
	500: "Server Errors	Something went wrong on Storyblok's end. (These are rare.)",
	502: "Server Errors	Something went wrong on Storyblok's end. (These are rare.)",
	503: "Server Errors	Something went wrong on Storyblok's end. (These are rare.)",
	504: "Server Errors	Something went wrong on Storyblok's end. (These are rare.)",
}
