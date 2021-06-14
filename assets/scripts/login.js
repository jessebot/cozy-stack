;(function (w, d) {
  if (!w.fetch || !w.Headers) return

  const loginForm = d.getElementById('login-form')
  const passphraseInput = d.getElementById('password')
  const submitButton = d.getElementById('login-submit')
  const redirectInput = d.getElementById('redirect')
  const csrfTokenInput = d.getElementById('csrf_token')
  const stateInput = d.getElementById('state')
  const clientIdInput = d.getElementById('client_id')
  const loginField = d.getElementById('login-field')
  const longRunCheckbox = d.getElementById('long-run-session')
  const trustedTokenInput = d.getElementById('trusted-device-token')

  // Set the trusted device token from the localstorage in the form if it exists
  try {
    const storage = w.localStorage
    const deviceToken = storage.getItem('trusted-device-token') || ''
    trustedTokenInput.value = deviceToken
  } catch (e) {
    // do nothing
  }

  let errorPanel = loginField.querySelector('.invalid-tooltip')
  const showError = function (message) {
    let error = 'The Cozy server is unavailable. Do you have network?'
    if (message) {
      error = '' + message
    }

    if (errorPanel) {
      errorPanel.lastChild.textContent = error
    } else {
      errorPanel = d.createElement('div')
      errorPanel.classList.add('invalid-tooltip', 'mb-1')
      const arrow = d.createElement('div')
      arrow.classList.add('tooltip-arrow')
      errorPanel.appendChild(arrow)
      const icon = d.createElement('span')
      icon.classList.add('icon', 'icon-alert', 'bg-danger')
      errorPanel.appendChild(icon)
      errorPanel.append(error)
      loginField.appendChild(errorPanel)
    }

    passphraseInput.classList.add('is-invalid')
    passphraseInput.select()
    submitButton.removeAttribute('disabled')
  }

  const onSubmitPassphrase = function (event) {
    event.preventDefault()
    submitButton.setAttribute('disabled', true)

    const passphrase = passphraseInput.value
    const longRun = longRunCheckbox && longRunCheckbox.checked ? '1' : '0'
    const redirect = redirectInput && redirectInput.value + w.location.hash

    let passPromise = Promise.resolve(passphrase)
    const salt = loginForm.dataset.salt
    const iterations = parseInt(loginForm.dataset.iterations, 10)
    if (iterations > 0) {
      passPromise = w.password
        .hash(passphrase, salt, iterations)
        .then(({ hashed }) => hashed)
    }

    passPromise
      .then((pass) => {
        const data = new URLSearchParams()
        data.append('passphrase', pass)
        data.append('trusted-device-token', trustedTokenInput.value)
        data.append('long-run-session', longRun)
        data.append('redirect', redirect)
        data.append('csrf_token', csrfTokenInput.value)

        // For the /auth/authorize/move && /auth/confirm pages
        if (stateInput) {
          data.append('state', stateInput.value)
        }
        if (clientIdInput) {
          data.append('client_id', clientIdInput.value)
        }

        const headers = new Headers()
        headers.append('Content-Type', 'application/x-www-form-urlencoded')
        headers.append('Accept', 'application/json')
        return fetch(loginForm.action, {
          method: 'POST',
          headers: headers,
          body: data,
          credentials: 'same-origin',
        })
      })
      .then((response) => {
        return response.json().then((body) => {
          if (response.status < 400) {
            submitButton.innerHTML = '<span class="icon icon-check"></span>'
            submitButton.classList.add('btn-done')
            w.location = body.redirect
          } else {
            showError(body.error)
          }
        })
      })
      .catch(showError)
  }

  loginForm.addEventListener('submit', onSubmitPassphrase)
  passphraseInput.focus()
  submitButton.removeAttribute('disabled')
})(window, document)
