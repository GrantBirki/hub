# hub tab-completion script for bash.
# This script complements the completion script that ships with git.

# If there is no git tab completion, but we have the _completion loader try to load it
if ! declare -F _git > /dev/null && declare -F _completion_loader > /dev/null; then
  _completion_loader git
fi

# Check that git tab completion is available and we haven't already set up completion
if declare -F _git > /dev/null && ! declare -F __git_list_all_commands_without_hub > /dev/null; then
  # Duplicate and rename the 'list_all_commands' function
  eval "$(declare -f __git_list_all_commands | \
        sed 's/__git_list_all_commands/__git_list_all_commands_without_hub/')"

  # Wrap the 'list_all_commands' function with extra hub commands
  __git_list_all_commands() {
    cat <<-EOF
alias
pull-request
pr
issue
release
fork
create
delete
browse
compare
ci-status
sync
EOF
    __git_list_all_commands_without_hub
  }

  # Ensure cached commands are cleared
  __git_all_commands=""

  ##########################
  # hub command completions
  ##########################

  # hub alias [-s] [SHELL]
  _git_alias() {
    local i c=2 s=-s sh shells="bash zsh sh ksh csh fish"
    while [ $c -lt $cword ]; do
      i="${words[c]}"
      case "$i" in
        -s)
          unset s
          ;;
        *)
          for sh in $shells; do
            if [ "$sh" = "$i" ]; then
              unset shells
              break
            fi
          done
          ;;
      esac
      ((c++))
    done
    __gitcomp "$s $shells"
  }

  # hub browse [-u] [--|[USER/]REPOSITORY] [SUBPAGE]
  _git_browse() {
    local i c=2 u=-u repo subpage
    local subpages_="commits issues tree wiki pulls branches stargazers
      contributors network network/ graphs graphs/"
    local subpages_network="members"
    local subpages_graphs="commit-activity code-frequency punch-card"
    while [ $c -lt $cword ]; do
      i="${words[c]}"
      case "$i" in
        -u)
          unset u
          ;;
        *)
          if [ -z "$repo" ]; then
            repo=$i
          else
            subpage=$i
          fi
          ;;
      esac
      ((c++))
    done
    if [ -z "$repo" ]; then
      __gitcomp "$u -- $(__hub_github_repos '\p')"
    elif [ -z "$subpage" ]; then
      case "$cur" in
        */*)
          local pfx="${cur%/*}" cur_="${cur#*/}"
          local subpages_var="subpages_$pfx"
          __gitcomp "${!subpages_var}" "$pfx/" "$cur_"
          ;;
        *)
          __gitcomp "$u ${subpages_}"
          ;;
      esac
    else
      __gitcomp "$u"
    fi
  }

  # hub compare [-u] [USER[/REPOSITORY]] [[START...]END]
  _git_compare() {
    local i c=$((cword - 1)) u=-u user remote owner repo arg_repo rev
    while [ $c -gt 1 ]; do
      i="${words[c]}"
      case "$i" in
        -u)
          unset u
          ;;
        *)
          if [ -z "$rev" ]; then
            # Even though the logic below is able to complete both user/repo
            # and revision in the right place, when there is only one argument
            # (other than -u) in the command, that argument will be taken as
            # revision. For example:
            # $ hub compare -u upstream
            # > https://github.com/USER/REPO/compare/upstream
            if __hub_github_repos '\p' | command grep -Eqx "^$i(/[^/]+)?"; then
              arg_repo=$i
            else
              rev=$i
            fi
          elif [ -z "$arg_repo" ]; then
            arg_repo=$i
          fi
          ;;
      esac
      ((c--))
    done

    # Here we want to find out the git remote name of user/repo, in order to
    # generate an appropriate revision list
    if [ -z "$arg_repo" ]; then
      user=$(__hub_github_user)
      if [ -z "$user" ]; then
        for i in $(__hub_github_repos); do
          remote=${i%%:*}
          repo=${i#*:}
          if [ "$remote" = origin ]; then
            break
          fi
        done
      else
        for i in $(__hub_github_repos); do
          remote=${i%%:*}
          repo=${i#*:}
          owner=${repo%%/*}
          if [ "$user" = "$owner" ]; then
            break
          fi
        done
      fi
    else
      for i in $(__hub_github_repos); do
        remote=${i%%:*}
        repo=${i#*:}
        owner=${repo%%/*}
        case "$arg_repo" in
          "$repo"|"$owner")
            break
            ;;
        esac
      done
    fi

    local pfx cur_="$cur"
    case "$cur_" in
      *..*)
        pfx="${cur_%%..*}..."
        cur_="${cur_##*..}"
        __gitcomp_nl "$(__hub_revlist $remote)" "$pfx" "$cur_"
        ;;
      *)
        if [ -z "${arg_repo}${rev}" ]; then
          __gitcomp "$u $(__hub_github_repos '\o\n\p') $(__hub_revlist $remote)"
        elif [ -z "$rev" ]; then
          __gitcomp "$u $(__hub_revlist $remote)"
        else
          __gitcomp "$u"
        fi
        ;;
    esac
  }

  # hub create [NAME] [-p] [-d DESCRIPTION] [-h HOMEPAGE]
  _git_create() {
    local i c=2 name repo flags="-p -d -h"
    while [ $c -lt $cword ]; do
      i="${words[c]}"
      case "$i" in
        -d|-h)
          ((c++))
          flags=${flags/$i/}
          ;;
        -p)
          flags=${flags/$i/}
          ;;
        *)
          name=$i
          ;;
      esac
      ((c++))
    done
    if [ -z "$name" ]; then
      repo=$(basename "$(pwd)")
    fi
    case "$prev" in
      -d|-h)
        COMPREPLY=()
        ;;
      -p|*)
        __gitcomp "$repo $flags"
        ;;
    esac
  }

  # hub fork [--no-remote] [--remote-name REMOTE] [--org ORGANIZATION]
  _git_fork() {
    local i c=2 flags="--no-remote --remote-name --org"
    while [ $c -lt $cword ]; do
      i="${words[c]}"
      case "$i" in
        --org)
          ((c++))
          flags=${flags/$i/}
          ;;
        --remote-name)
          ((c++))
          flags=${flags/$i/}
          flags=${flags/--no-remote/}
          ;;
        --no-remote)
          flags=${flags/$i/}
          flags=${flags/--remote-name/}
          ;;
      esac
      ((c++))
    done
    case "$prev" in
      --remote-name|--org)
        COMPREPLY=()
        ;;
      *)
        __gitcomp "$flags"
        ;;
    esac
  }

  # hub pull-request [-f] [-m <MESSAGE>|-F <FILE>|-i <ISSUE>|<ISSUE-URL>] [-b <BASE>] [-h <HEAD>] [-a <USER>] [-M <MILESTONE>] [-l <LABELS>]
  _git_pull_request() {
    local i c=2 flags="-f -m -F -i -b -h -a -M -l"
    while [ $c -lt $cword ]; do
      i="${words[c]}"
      case "$i" in
        -m|-F|-i|-b|-h|-a|-M|-l)
          ((c++))
          flags=${flags/$i/}
          ;;
        -f)
          flags=${flags/$i/}
          ;;
      esac
      ((c++))
    done
    case "$prev" in
      -i)
        COMPREPLY=()
        ;;
      -b|-h|-a|-M|-l)
        # (Doesn't seem to need this...)
        # Uncomment the following line when 'owner/repo:[TAB]' misbehaved
        #_get_comp_words_by_ref -n : cur
        __gitcomp_nl "$(__hub_heads)"
        # __ltrim_colon_completions "$cur"
        ;;
      -F)
        COMPREPLY=( "$cur"* )
        ;;
      -f|*)
        __gitcomp "$flags"
        ;;
    esac
  }

  ###################
  # Helper functions
  ###################

  # __hub_github_user [HOST]
  # Return $GITHUB_USER or the default github user defined in hub config
  # HOST - Host to be looked-up in hub config. Default is "github.com"
  __hub_github_user() {
    if [ -n "$GITHUB_USER" ]; then
      echo $GITHUB_USER
      return
    fi
    local line h k v host=${1:-github.com} config=${HUB_CONFIG:-~/.config/hub}
    if [ -f "$config" ]; then
      while read line; do
        if [ "$line" = "---" ]; then
          continue
        fi
        k=${line%%:*}
        v=${line#*:}
        if [ -z "$v" ]; then
          if [ "$h" = "$host" ]; then
            break
          fi
          h=$k
          continue
        fi
        k=${k#* }
        v=${v#* }
        if [ "$h" = "$host" ] && [ "$k" = "user" ]; then
          echo "$v"
          break
        fi
      done < "$config"
    fi
  }

  # __hub_github_repos [FORMAT]
  # List all github hosted repository
  # FORMAT - Format string contains multiple of these:
  #   \m  remote
  #   \p  owner/repo
  #   \o  owner
  #   escaped characters (\n, \t ...etc) work
  # If omitted, prints all github repos in the format of "remote:owner/repo"
  __hub_github_repos() {
    local f format=$1
    if [ -z "$(__gitdir)" ]; then
      return
    fi
    if [ -z "$format" ]; then
      format='\1:\2'
    else
      format=${format//\m/\1}
      format=${format//\p/\2}
      format=${format//\o/\3}
    fi
    command git config --get-regexp 'remote\.[^.]*\.url' |
    command grep -E ' ((https?|git)://|git@)github\.com[:/][^:/]+/[^/]+$' |
    sed -E 's#^remote\.([^.]+)\.url +.+[:/](([^/]+)/[^.]+)(\.git)?$#'"$format"'#'
  }

  # __hub_heads
  # List all local "branch", and remote "owner/repo:branch"
  __hub_heads() {
    local i remote repo branch dir=$(__gitdir)
    if [ -d "$dir" ]; then
      command git --git-dir="$dir" for-each-ref --format='%(refname:short)' \
        "refs/heads/"
      for i in $(__hub_github_repos); do
        remote=${i%%:*}
        repo=${i#*:}
        command git --git-dir="$dir" for-each-ref --format='%(refname:short)' \
          "refs/remotes/${remote}/" | while read branch; do
          echo "${repo}:${branch#${remote}/}"
        done
      done
    fi
  }

  # __hub_revlist [REMOTE]
  # List all tags, and branches under REMOTE, without the "remote/" prefix
  # REMOTE - Remote name to search branches from. Default is "origin"
  __hub_revlist() {
    local i remote=${1:-origin} dir=$(__gitdir)
    if [ -d "$dir" ]; then
      command git --git-dir="$dir" for-each-ref --format='%(refname:short)' \
        "refs/remotes/${remote}/" | while read i; do
        echo "${i#${remote}/}"
      done
      command git --git-dir="$dir" for-each-ref --format='%(refname:short)' \
        "refs/tags/"
    fi
  }

  __hub_main() {
    local can_unalias

    if [ "$(type -t git)" != alias ]; then
      alias git=hub
      can_unalias=true
    fi

    __git_main "$@"

    if [ -n "$can_unalias" ]; then
      unalias git
    fi
  }

  # Enable completion for hub even when not using the alias
  __git_complete hub __hub_main
fi
