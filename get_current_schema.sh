#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

readonly working_dir=$(dirname "$(readlink --canonicalize-existing "${0}")")
readonly conf_file="${working_dir}/getschema.conf"
readonly error_reading_conf_file=80
readonly retries_to_get_schema=5
readonly seconds_to_wait_on_failure_getting_ds=30

. "${conf_file}" || {
    echo "error reading conf file" >&2
    exit ${error_reading_conf_file}
}

# ORACLE_HOME="/az/oracle/db/product/12.1.0/client_1"
export DB_CONNECTION=AZO_CORE/atgdbpwd@pr-camaro1-sc01.autozone.com:19100/az_atg_diy.autozone.com

cd "${working_dir}"

get_current_aux_schema() {
    # Variable to control the retry loop:
    local i
    local got_schema=0

    rm --force "${working_dir}/current_schema"

    for ((i = 0; i < retries_to_get_schema; i++)); do
        "${ORACLE_HOME}/bin/sqlplus" -s "${DB_CONNECTION}" <<SQL_QUERY > "${working_dir}/current_schema" 2>&1
        set heading off
        set feedback off
        set pages 0
        select curr_ds_name from das_sds where sds_name = 'AuxiliarySwitchingDataSource';
SQL_QUERY

        LC_ALL=C grep --extended-regexp --quiet '^AuxDataSource[A|B]$' "${working_dir}/current_schema" && {
            got_schema=1
            break
        }
        sleep "${seconds_to_wait_on_failure_getting_ds:-30}"s

    done

    if ((got_schema == 1)); then
        cat "${working_dir}/current_schema"
    else
        # Return empty or exit ...
        echo ""
    fi
}

exit 0
