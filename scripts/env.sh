load_dotenv() {
  local file="${1:-.env}"
  local line key value

  [[ -r "$file" ]] || {
    printf 'Cannot read environment file: %s\n' "$file" >&2
    return 1
  }

  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line%$'\r'}"
    [[ "$line" =~ ^[[:space:]]*$ ]] && continue
    [[ "$line" =~ ^[[:space:]]*# ]] && continue
    [[ "$line" == *=* ]] || {
      printf 'Invalid environment line in %s\n' "$file" >&2
      return 1
    }

    key="${line%%=*}"
    value="${line#*=}"
    key="${key//[[:space:]]/}"
    [[ "$key" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || {
      printf 'Invalid environment variable name in %s\n' "$file" >&2
      return 1
    }

    if [[ "$value" == \"*\" && "$value" == *\" ]]; then
      value="${value:1:${#value}-2}"
    elif [[ "$value" == \'*\' && "$value" == *\' ]]; then
      value="${value:1:${#value}-2}"
    fi

    export "$key=$value"
  done < "$file"
}
