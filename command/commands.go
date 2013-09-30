/*
   Copyright 2013 Matthew Collins (purggames@gmail.com)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package command

func init() {
	Register("echo #string", command_echo)
	Register("echo #number", command_echoNumber)
}

func command_echo(caller Caller, msg string) (string, string) {
	return msg, ""
}

func command_echoNumber(caller Caller, msg float64) (float64, string) {
	return msg, ""
}
