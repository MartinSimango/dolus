[
	{
		"path":     "/"
		"method":   "GET"
		"priority": 1
		"request": {
			"body":    "a"
			"headers": "b"
		}

		"response": {
			"body": {
				"list": [
					{
						"a": "b" // DOLUS_UUID()
					},
				]
			}
			"headers": "b"
		}

	},
]
