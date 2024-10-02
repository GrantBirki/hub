Feature: hub fork
  Background:
    Given I am in "dotfiles" git repo
    And the "origin" remote has url "git://github.com/evilchelu/dotfiles.git"
    And I am "mislav" on github.com with OAuth token "OTOKEN"

  Scenario: Fork the repository
    Given the GitHub API server:
      """
      before {
        halt 400 unless request.env['HTTP_X_ORIGINAL_SCHEME'] == 'https'
        halt 401 unless request.env['HTTP_AUTHORIZATION'] == 'token OTOKEN'
      }
      get('/repos/mislav/dotfiles', :host_name => 'api.github.com') { 404 }
      post('/repos/evilchelu/dotfiles/forks', :host_name => 'api.github.com') {
        assert :organization => nil
        status 202
        json :name => 'dotfiles', :owner => { :login => 'mislav' }
      }
      """
    When I successfully run `hub fork`
    Then the output should contain exactly "new remote: mislav\n"
    And "git remote add -f mislav https://github.com/evilchelu/dotfiles.git" should be run
    And "git remote set-url mislav https://github.com/mislav/dotfiles.git" should be run
    And the url for "mislav" should be "https://github.com/mislav/dotfiles.git"

  Scenario: Fork the repository with new remote name specified
    Given git protocol is preferred
    Given the GitHub API server:
      """
      get('/repos/mislav/dotfiles') { 404 }
      post('/repos/evilchelu/dotfiles/forks') {
        assert :organization => nil
        status 202
        json :name => 'dotfiles', :owner => { :login => 'mislav' }
      }
      """
    When I successfully run `hub fork --remote-name=origin`
    Then the output should contain exactly:
      """
      renaming existing "origin" remote to "upstream"
      new remote: origin\n
      """
    And "git remote add -f origin git://github.com/evilchelu/dotfiles.git" should be run
    And "git remote set-url origin git@github.com:mislav/dotfiles.git" should be run
    And the url for "origin" should be "git@github.com:mislav/dotfiles.git"
    And the url for "upstream" should be "git://github.com/evilchelu/dotfiles.git"

  Scenario: Fork the repository with redirect
    Given git protocol is preferred
    Given the GitHub API server:
      """
      before {
        halt 400 unless request.env['HTTP_X_ORIGINAL_SCHEME'] == 'https'
        halt 401 unless request.env['HTTP_AUTHORIZATION'] == 'token OTOKEN'
      }
      get('/repos/mislav/dotfiles', :host_name => 'api.github.com') { 404 }
      post('/repos/evilchelu/dotfiles/forks', :host_name => 'api.github.com') {
        redirect 'https://api.github.com/repositories/1234/forks', 307
      }
      post('/repositories/1234/forks', :host_name => 'api.github.com') {
        status 202
        json :name => 'my-dotfiles', :owner => { :login => 'MiSlAv' }
      }
      """
    When I successfully run `hub fork`
    Then the output should contain exactly "new remote: mislav\n"
    And "git remote add -f mislav git://github.com/evilchelu/dotfiles.git" should be run
    And "git remote set-url mislav git@github.com:MiSlAv/my-dotfiles.git" should be run
    And the url for "mislav" should be "git@github.com:MiSlAv/my-dotfiles.git"

  Scenario: Fork the repository when origin URL is private
    Given the "origin" remote has url "git@github.com:evilchelu/dotfiles.git"
    And git protocol is preferred
    Given the GitHub API server:
      """
      before { halt 401 unless request.env['HTTP_AUTHORIZATION'] == 'token OTOKEN' }
      get('/repos/mislav/dotfiles', :host_name => 'api.github.com') { 404 }
      post('/repos/evilchelu/dotfiles/forks', :host_name => 'api.github.com') {
        status 202
        json :name => 'dotfiles', :owner => { :login => 'mislav' }
      }
      """
    When I successfully run `hub fork`
    Then the output should contain exactly "new remote: mislav\n"
    And "git remote add -f mislav git://github.com/evilchelu/dotfiles.git" should be run
    And "git remote set-url mislav git@github.com:mislav/dotfiles.git" should be run
    And the url for "mislav" should be "git@github.com:mislav/dotfiles.git"

  Scenario: --no-remote
    Given the GitHub API server:
      """
      post('/repos/evilchelu/dotfiles/forks') {
        status 202
        json :name => 'dotfiles', :owner => { :login => 'mislav' }
      }
      """
    When I successfully run `hub fork --no-remote`
    Then the output should not contain anything
    And there should be no "mislav" remote

  Scenario: --set-push-default
    Given the GitHub API server:
      """
      post('/repos/evilchelu/dotfiles/forks') {
        status 202
        json :name => 'dotfiles', :owner => { :login => 'mislav' }
      }
      """
    When I successfully run `hub fork --set-push-default`
    Then the "remote.pushDefault" config option should be "mislav"

  Scenario: Fork failed
    Given the GitHub API server:
      """
      post('/repos/evilchelu/dotfiles/forks') { halt 500 }
      """
    When I run `hub fork`
    Then the exit status should be 1
    And the stderr should contain exactly:
      """
      Error creating fork: Internal Server Error (HTTP 500)\n
      """
    And there should be no "mislav" remote

  Scenario: Unrelated fork already exists
    Given the GitHub API server:
      """
      get('/repos/mislav/dotfiles') {
        halt 406 unless request.env['HTTP_ACCEPT'] == 'application/vnd.github.v3+json;charset=utf-8'
        json :html_url => 'https://github.com/mislav/dotfiles',
             :parent => { :html_url => 'https://github.com/unrelated/dotfiles' }
      }
      """
    When I run `hub fork`
    Then the exit status should be 1
    And the stderr should contain exactly:
      """
      Error creating fork: mislav/dotfiles already exists on github.com\n
      """
    And there should be no "mislav" remote

  Scenario: Related fork already exists
    Given the GitHub API server:
      """
      get('/repos/mislav/dotfiles') {
        json :html_url => 'https://github.com/mislav/dotfiles',
             :parent => { :html_url => 'https://github.com/EvilChelu/Dotfiles' }
      }
      """
    When I run `hub fork`
    Then the exit status should be 0
    Then the stdout should contain exactly:
      """
      new remote: mislav\n
      """
    And the url for "mislav" should be "https://github.com/mislav/dotfiles.git"

  Scenario: Redirected repo already exists
    Given the GitHub API server:
      """
      get('/repos/mislav/dotfiles') {
        redirect 'https://api.github.com/repositories/12345', 301
      }
      get('/repositories/12345') {
        json :html_url => 'https://github.com/mislav/old-dotfiles'
      }
      post('/repos/evilchelu/dotfiles/forks') {
        status 202
        json :name => 'dotfiles', :owner => { :login => 'mislav' }
      }
      """
    When I successfully run `hub fork`
    And the stdout should contain exactly "new remote: mislav\n"

  Scenario: Unrelated remote already exists
    Given the "mislav" remote has url "git@github.com:mislav/unrelated.git"
    Given the GitHub API server:
      """
      get('/repos/mislav/dotfiles', :host_name => 'api.github.com') { 404 }
      post('/repos/evilchelu/dotfiles/forks', :host_name => 'api.github.com') {
        assert :organization => nil
        status 202
        json :name => 'dotfiles', :owner => { :login => 'mislav' }
      }
      """
    When I run `hub fork`
    Then the exit status should not be 0
    And the stderr should contain:
      """
      remote mislav already exists.
      """
    And the url for "mislav" should be "git@github.com:mislav/unrelated.git"

  Scenario: Related fork and related remote already exist
    Given the "mislav" remote has url "git@github.com:mislav/dotfiles.git"
    Given the GitHub API server:
      """
      get('/repos/mislav/dotfiles') {
        json :html_url => 'https://github.com/mislav/dotfiles',
             :parent => { :html_url => 'https://github.com/EvilChelu/Dotfiles' }
      }
      """
    When I run `hub fork`
    Then the exit status should be 0
    And the stdout should contain exactly:
      """
      existing remote: mislav\n
      """
    And the url for "mislav" should be "git@github.com:mislav/dotfiles.git"

  Scenario: Related fork and related remote, but with differing protocol, already exist
    Given the "mislav" remote has url "https://github.com/mislav/dotfiles.git"
    Given the GitHub API server:
      """
      get('/repos/mislav/dotfiles') {
        json :html_url => 'https://github.com/mislav/dotfiles',
             :parent => { :html_url => 'https://github.com/EvilChelu/Dotfiles' }
      }
      """
    When I run `hub fork`
    Then the exit status should be 0
    And the stdout should contain exactly:
      """
      existing remote: mislav\n
      """
    And the url for "mislav" should be "https://github.com/mislav/dotfiles.git"

  Scenario: Invalid OAuth token
    Given the GitHub API server:
      """
      before { halt 401 unless request.env['HTTP_AUTHORIZATION'] == 'token OTOKEN' }
      """
    And I am "mislav" on github.com with OAuth token "WRONGTOKEN"
    When I run `hub fork`
    Then the exit status should be 1
    And the stderr should contain exactly:
      """
      Error creating fork: Unauthorized (HTTP 401)\n
      """

  Scenario: Not in repo
    Given the current dir is not a repo
    When I run `hub fork`
    Then the exit status should be 1
    And the stderr should contain "fatal: Not a git repository"

  Scenario: Origin remote doesn't exist
    Given I run `git remote rm origin`
    When I run `hub fork`
    Then the exit status should be 1
    And the stderr should contain exactly:
      """
      Aborted: could not find any git remote pointing to a GitHub repository\n
      """
    And there should be no "origin" remote

  Scenario: Unknown host
    Given the "origin" remote has url "git@git.my.org:evilchelu/dotfiles.git"
    When I run `hub fork`
    Then the exit status should be 1
    And the stderr should contain exactly:
      """
      Aborted: could not find any git remote pointing to a GitHub repository\n
      """

  Scenario: Enterprise fork
    Given the GitHub API server:
      """
      before {
        halt 400 unless request.env['HTTP_X_ORIGINAL_SCHEME'] == 'https'
        halt 401 unless request.env['HTTP_AUTHORIZATION'] == 'token FITOKEN'
      }
      post('/api/v3/repos/evilchelu/dotfiles/forks', :host_name => 'git.my.org') {
        status 202
        json :name => 'dotfiles', :owner => { :login => 'mislav' }
      }
      """
    And the "origin" remote has url "git@git.my.org:evilchelu/dotfiles.git"
    And I am "mislav" on git.my.org with OAuth token "FITOKEN"
    And "git.my.org" is a whitelisted Enterprise host
    When I successfully run `hub fork`
    Then the url for "mislav" should be "https://git.my.org/mislav/dotfiles.git"

  Scenario: Enterprise fork using regular HTTP
    Given the GitHub API server:
      """
      before {
        halt 400 unless request.env['HTTP_X_ORIGINAL_SCHEME'] == 'http'
        halt 400 unless request.env['HTTP_X_ORIGINAL_PORT'] == '80'
        halt 401 unless request.env['HTTP_AUTHORIZATION'] == 'token FITOKEN'
      }
      post('/api/v3/repos/evilchelu/dotfiles/forks', :host_name => 'git.my.org') {
        status 202
        json :name => 'dotfiles', :owner => { :login => 'mislav' }
      }
      """
    And the "origin" remote has url "git@git.my.org:evilchelu/dotfiles.git"
    And I am "mislav" on http://git.my.org with OAuth token "FITOKEN"
    And "git.my.org" is a whitelisted Enterprise host
    When I successfully run `hub fork`
    Then the url for "mislav" should be "https://git.my.org/mislav/dotfiles.git"

  Scenario: Fork a repo to a specific organization
    Given the GitHub API server:
      """
      get('/repos/acme/dotfiles') { 404 }
      post('/repos/evilchelu/dotfiles/forks') {
        assert :organization => "acme"
        status 202
        json :name => 'dotfiles', :owner => { :login => 'acme' }
      }
      """
    When I successfully run `hub fork --org=acme`
    Then the output should contain exactly "new remote: acme\n"
    Then the url for "acme" should be "https://github.com/acme/dotfiles.git"
