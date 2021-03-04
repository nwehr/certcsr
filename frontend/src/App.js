import React, { useState } from "react"
import { Form, Button, Alert } from "react-bootstrap"
import axios from "axios"

import 'bootstrap/dist/css/bootstrap.min.css';
import './App.css';

function RequestForm() {
  const [password, setPassword] = useState(null)

  const [country, setCountry] = useState("US")
  const [state, setState] = useState("Virginia")
  const [city, setCity] = useState("Roanoke")
  const [company, setCompany] = useState("Wehr Holdings, LLC")
  const [name, setName] = useState("")
  const [email, setEmail] = useState("")

  const [submitWaiting, setSubmitWaiting] = useState(false)

  const submit = async (e) => {
	e.preventDefault()

	setSubmitWaiting(true)

	try {
	  const { status, data } = await axios.post("/post-csr", { country, state, city, company, name, email })
	  if (status == 200) {
		setPassword(data.password)
	  }
	}
	catch (e) {
	  alert(e)
	}

	setSubmitWaiting(false)
  }

  if (password != null) {
	return <Alert variant="dark" style={{fontFamily:"courier new; courier"}}>{password}</Alert>
  }

  return <Form onSubmit={submit}>
	<Form.Group>
	  <Form.Label>Country</Form.Label>
	  <Form.Control value={country} onChange={e => setCountry(e.target.value)} />
	</Form.Group>
	<Form.Group>
	  <Form.Label>State</Form.Label>
	  <Form.Control value={state} onChange={e => setState(e.target.value)} />
	</Form.Group>
	<Form.Group>
	  <Form.Label>City</Form.Label>
	  <Form.Control value={city} onChange={e => setCity(e.target.value)} />
	</Form.Group>
	<Form.Group>
	  <Form.Label>Company</Form.Label>
	  <Form.Control value={company} onChange={e => setCompany(e.target.value)} />
	</Form.Group>
	<Form.Group>
	  <Form.Label>Full Name</Form.Label>
	  <Form.Control value={name} onChange={e => setName(e.target.value)} />
	</Form.Group>
	<Form.Group>
	  <Form.Label>Email</Form.Label>
	  <Form.Control value={email} onChange={e => setEmail(e.target.value)} />
	</Form.Group>
	<Button variant="primary" type="submit" disabled={submitWaiting}>{submitWaiting ? "Submit..." : "Submit"}</Button>
  </Form>
}

function App() {
  return (
	<div className="App">
	  <RequestForm />
	</div>
  );
}

export default App;
