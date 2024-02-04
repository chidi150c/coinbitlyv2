package model

type AIAgent interface{
	AIModelAPI(string)(string, error)
}