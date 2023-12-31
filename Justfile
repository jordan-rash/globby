just := env_var_or_default("JUST", just_executable())
wasm_tools := env_var_or_default("WASM_TOOLS", "wasm-tools")
wash := env_var_or_default("WASH", "wash")
tinygo := env_var_or_default("TINYGO", "tinygo")
wit_bindgen := env_var_or_default("WIT_BINDGEN", "wit-bindgen")

actorid := env_var_or_default("ACTOR_ID", "")
hostid := env_var_or_default("HOST_ID", "")

module_name := "globby"
expected_wasm_path := "./build/globby.wasm"
expected_wasm_embed_path := "./build/globby.embed.wasm"
expected_wasm_component_path := "./build/globby.component.wasm"
wasm_preview2_output_path := "./wasi_snapshot_preview1.command.wasm"

_default:
  {{just}} --list

@versions:
  {{tinygo}} version
  {{wash}} --version
  {{wasm_tools}} --version
  {{wit_bindgen}} --version


build:
  mkdir -p build
  {{tinygo}} build -target=wasi -scheduler=none -o {{expected_wasm_path}} .
  {{wash}} claims sign --name {{module_name}} {{expected_wasm_path}} -f -q -l

clean:
  rm -rf ./build

signed_actor_path := absolute_path("./build/kvcounter-wasi.component_s.wasm")
start_actor:
  {{wash}} start actor file://{{signed_actor_path}} --host-id {{hostid}}

restart_actor:
  {{wash}} stop actor {{hostid}} {{actorid}}
  {{wash}} drain lib
  {{wash}} start actor file://{{signed_actor_path}} --host-id {{hostid}}
