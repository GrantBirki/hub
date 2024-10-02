require 'fileutils'

Given(/^git protocol is preferred$/) do
  set_environment_variable "HUB_PROTOCOL", "git"
end

Given(/^there are no remotes$/) do
  output = run_ignored_command 'git remote'
  expect(output).to be_empty
end

Given(/^"([^"]*)" is a whitelisted Enterprise host$/) do |host|
  run_ignored_command %(git config --global --add hub.host "#{host}")
end

Given(/^git "(.+?)" is set to "(.+?)"$/) do |key, value|
  run_ignored_command %(git config #{key} "#{value}")
end

Given(/^the "([^"]*)" remote has(?: (push))? url "([^"]*)"$/) do |remote_name, push, url|
  remotes = run_ignored_command 'git remote'
  unless remotes.split("\n").include? remote_name
    run_ignored_command %(git remote add #{remote_name} "#{url}")
  else
    run_ignored_command %(git remote set-url #{"--push" if push} #{remote_name} "#{url}")
  end
end

Given(/^I am "([^"]*)" on ([\S]+)(?: with OAuth token "([^"]*)")?$/) do |name, host, token|
  edit_hub_config do |cfg|
    entry = {'user' => name}
    host = host.sub(%r{^([\w-]+)://}, '')
    entry['oauth_token'] = token if token
    entry['protocol'] = $1 if $1
    cfg[host.downcase] = [entry]
  end
end

Given(/^\$(\w+) is "([^"]*)"$/) do |name, value|
  expanded_value = value.gsub(/\$([A-Z_]+)/) { aruba.environment[$1] }
  set_environment_variable(name, expanded_value)
end

Given(/^I am in "([^"]*)" git repo$/) do |dir_name|
  if dir_name.include?(':')
    origin_url = dir_name
    dir_name = File.basename origin_url, '.git'
  end
  step %(a git repo in "#{dir_name}")
  step %(I cd to "#{dir_name}")
  step %(the "origin" remote has url "#{origin_url}") if origin_url
end

Given(/^a (bare )?git repo in "([^"]*)"$/) do |bare, dir_name|
  run_ignored_command %(git -c init.defaultBranch=master init --quiet #{"--bare" if bare} '#{dir_name}')
end

Given(/^a git bundle named "([^"]*)"$/) do |file|
  dest = expand_path(file)
  FileUtils.mkdir_p(File.dirname(dest))

  Dir.mktmpdir do |tmpdir|
    Dir.chdir(tmpdir) do
      `git -c init.defaultBranch=master init --quiet`
      `GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=b git commit --quiet -m 'empty' --allow-empty --author='a <b>'`
      `git bundle create "#{dest}" master 2>&1`
    end
  end
end

Given(/^there is a commit named "([^"]+)"$/) do |name|
  empty_commit
  empty_commit
  run_ignored_command %(git tag #{name})
  run_ignored_command %(git reset --quiet --hard HEAD^)
end

Given(/^there is a git FETCH_HEAD$/) do
  empty_commit
  empty_commit
  cd('.') do
    File.open(".git/FETCH_HEAD", "w") do |fetch_head|
      fetch_head.puts "%s\t\t'refs/heads/made-up' of git://github.com/made/up.git" % `git rev-parse HEAD`.chomp
    end
  end
  run_ignored_command %(git reset --quiet --hard HEAD^)
end

When(/^I make (a|\d+) commits?(?: with message "([^"]+)")?$/) do |num, msg|
  num = num == 'a' ? 1 : num.to_i
  num.times { empty_commit(msg) }
end

When(/^I make a commit with message:$/) do |msg|
  empty_commit(msg)
end

Then(/^the latest commit message should be "([^"]+)"$/) do |subject|
  step %(I successfully run `git log -1 --format=%s`)
  step %(the output should contain exactly "#{subject}\\n")
end

# expand `<$HOME>` etc. in matched text
Then(/^(the (?:output|stderr|stdout)) with expanded variables( should contain(?: exactly)?:)/) do |prefix, postfix, text|
  step %(#{prefix}#{postfix}), text.gsub(/<\$(\w+)>/) { aruba.environment[$1] }
end

Given(/^the "([^"]+)" branch is pushed to "([^"]+)"$/) do |name, upstream|
  full_upstream = ".git/refs/remotes/#{upstream}"
  cd('.') do
    FileUtils.mkdir_p File.dirname(full_upstream)
    FileUtils.cp ".git/refs/heads/#{name}", full_upstream
  end
end

Given(/^I am on the "([^"]+)" branch(?: (pushed to|with upstream) "([^"]+)")?$/) do |name, type, upstream|
  run_ignored_command %(git checkout --quiet -b #{shell_escape name})
  empty_commit

  if upstream
    full_upstream = upstream.start_with?('refs/') ? upstream : "refs/remotes/#{upstream}"
    run_ignored_command %(git update-ref #{shell_escape full_upstream} HEAD)

    if type == 'with upstream'
      run_ignored_command %(git branch --set-upstream-to #{shell_escape upstream})
    end
  end
end

Given(/^the default branch for "([^"]+)" is "([^"]+)"$/) do |remote, branch|
  cd('.') do
    ref_file = ".git/refs/remotes/#{remote}/#{branch}"
    unless File.exist? ref_file
      empty_commit unless File.exist? '.git/refs/heads/master'
      FileUtils.mkdir_p File.dirname(ref_file)
      FileUtils.cp '.git/refs/heads/master', ref_file
    end
  end
  run_ignored_command %(git remote set-head #{remote} #{branch})
end

Given(/^I am in detached HEAD$/) do
  empty_commit
  empty_commit
  run_ignored_command %(git checkout HEAD^)
end

Given(/^the current dir is not a repo$/) do
  FileUtils.rm_rf(expand_path('.git'))
end

Given(/^the GitHub API server:$/) do |endpoints_str|
  @server = Hub::LocalServer.start_sinatra do
    eval endpoints_str, binding
  end
  # hit our Sinatra server instead of github.com
  set_environment_variable 'HUB_TEST_HOST', "http://127.0.0.1:#{@server.port}"
end

Then(/^shell$/) do
  cd('.') do
    system '/bin/bash -i'
  end
end

Then(/^"([^"]*)" should be run$/) do |cmd|
  assert_command_run cmd
end

Then(/^it should clone "([^"]*)"$/) do |repo|
  step %("git clone #{repo}" should be run)
end

Then(/^it should not clone anything$/) do
  history.each { |h| expect(h).to_not match(/^git clone/) }
end

Then(/^"([^"]+)" should not be run$/) do |pattern|
  history.each { |h| expect(h).to_not include(pattern) }
end

Then(/^the git command should be unchanged$/) do
  expect(@commands).to_not be_empty
  assert_command_run @commands.last.sub(/^hub\b/, 'git')
end

Then(/^the "([^"]*)" config option should be "([^"]*)"$/) do |name, value|
  run_command_and_stop %(git config --get-all "#{name}")
  expect(last_command_started).to have_output(value)
end

Then(/^the url for "([^"]*)" should be "([^"]*)"$/) do |name, url|
  output = run_ignored_command %(git config --get-all remote.#{name}.url)
  expect(output).to include(url)
end

Then(/^the "([^"]*)" submodule url should be "([^"]*)"$/) do |name, url|
  output = run_ignored_command %(git config --get-all submodule."#{name}".url)
  expect(output).to include(url)
end

Then(/^"([^"]*)" should merge "([^"]*)" from remote "([^"]*)"$/) do |name, merge, remote|
  output = run_ignored_command %(git config --get-all branch.#{name}.remote)
  expect(output).to include(remote)

  output = run_ignored_command %(git config --get-all branch.#{name}.merge)
  expect(output).to include(merge)
end

Then(/^there should be no "([^"]*)" remote$/) do |remote_name|
  remotes = run_ignored_command 'git remote'
  expect(remotes.split("\n")).to_not include(remote_name)
end

Then(/^the file "([^"]*)" should have mode "([^"]*)"$/) do |file, expected_mode|
  mode = File.stat(expand_path(file)).mode
  expect(mode.to_s(8)).to match(/#{expected_mode}$/)
end

Given(/^the remote commit states of "(.*?)" "(.*?)" are:$/) do |proj, ref, json_value|
  if ref == 'HEAD'
    empty_commit
  end
  output = run_ignored_command %(git rev-parse #{ref})
  rev = output.chomp

  host, owner, repo = proj.split('/', 3)
  if repo.nil?
    repo = owner
    owner = host
    host = nil
  end

  status_endpoint = <<-EOS
    get('#{'/api/v3' if host}/repos/#{owner}/#{repo}/commits/#{rev}/status'#{", :host_name => '#{host}'" if host}) {
      json(#{json_value})
    }
    get('#{'/api/v3' if host}/repos/#{owner}/#{repo}/commits/#{rev}/check-runs'#{", :host_name => '#{host}'" if host}) {
      status 422
    }
    EOS
  step %{the GitHub API server:}, status_endpoint
end

Given(/^the remote commit state of "(.*?)" "(.*?)" is "(.*?)"$/) do |proj, ref, status|
  step %{the remote commit states of "#{proj}" "#{ref}" are:}, <<-EOS
    { :state => "#{status}",
      :statuses => [
        { :state => "#{status}",
          :context => "continuous-integration/travis-ci/push",
          :target_url => 'https://travis-ci.org/#{proj}/builds/1234567' }
      ]
    }
  EOS
end

Given(/^the remote commit state of "(.*?)" "(.*?)" is nil$/) do |proj, ref|
  step %{the remote commit states of "#{proj}" "#{ref}" are:},
    %({ :state => "pending", :statuses => [] })
end

Given(/^the text editor exits with error status$/) do
  text_editor_script "exit 1"
end

Given(/^the text editor adds:$/) do |text|
  text_editor_script <<-BASH
    file="$3"
    contents="$(cat "$file" 2>/dev/null || true)"
    { echo "#{text}"
      echo
      echo "$contents"
    } > "$file"
  BASH
end

When(/^I pass in:$/) do |input|
  type(input)
  close_input
end

Given(/^the git commit editor is "([^"]+)"$/) do |cmd|
  set_environment_variable('GIT_EDITOR', cmd)
end

Given(/^the SSH config:$/) do |config_lines|
  ssh_config = expand_path('~/.ssh/config')
  FileUtils.mkdir_p(File.dirname(ssh_config))
  File.open(ssh_config, 'w') {|f| f << config_lines }
end

Given(/^the SHAs and timestamps are normalized in "([^"]+)"$/) do |file|
  file = expand_path(file)
  contents = File.read(file)
  contents.gsub!(/[0-9a-f]{7} \(Hub, \d seconds? ago\)/, "SHA1SHA (Hub, 0 seconds ago)")
  File.open(file, "w") { |f| f.write(contents) }
end

Then(/^its (output|stderr|stdout) should( not)? contain( exactly)?:$/) do |channel, negated, exactly, expected|
  matcher = case channel.to_sym
            when :output
              :have_output
            when :stderr
              :have_output_on_stderr
            when :stdout
              :have_output_on_stdout
            end

  commands = [last_command_started]

  output_string_matcher = if exactly
                            :an_output_string_being_eq
                          else
                            :an_output_string_including
                          end

  if negated
    expect(commands).not_to include_an_object send(matcher, send(output_string_matcher, expected))
  else
    expect(commands).to include_an_object send(matcher, send(output_string_matcher, expected))
  end
end
