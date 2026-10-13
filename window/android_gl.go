//go:build android

package window

import (
	"github.com/akrck02/go-renderer/gles"
	"golang.org/x/mobile/gl"
)

// mobileGL implements go-renderer's gles.Functions over golang.org/x/mobile/gl, so the shared
// OpenGL ES renderers draw on Android. It lives here, not in go-renderer, so that go-renderer
// (which other programs build on) does not depend on golang.org/x/mobile.
type mobileGL struct{ context gl.Context }

func (mobile mobileGL) CreateShader(stage gles.Enum) gles.Shader {
	return gles.Shader(mobile.context.CreateShader(gl.Enum(stage)).Value)
}

func (mobile mobileGL) ShaderSource(shader gles.Shader, source string) {
	mobile.context.ShaderSource(gl.Shader{Value: uint32(shader)}, source)
}

func (mobile mobileGL) CompileShader(shader gles.Shader) {
	mobile.context.CompileShader(gl.Shader{Value: uint32(shader)})
}

func (mobile mobileGL) ShaderCompiled(shader gles.Shader) (bool, string) {
	handle := gl.Shader{Value: uint32(shader)}
	if mobile.context.GetShaderi(handle, gl.COMPILE_STATUS) != 0 {
		return true, ""
	}
	return false, mobile.context.GetShaderInfoLog(handle)
}

func (mobile mobileGL) DeleteShader(shader gles.Shader) {
	mobile.context.DeleteShader(gl.Shader{Value: uint32(shader)})
}

func (mobile mobileGL) CreateProgram() gles.Program {
	return gles.Program(mobile.context.CreateProgram().Value)
}

// program names a program the way x/mobile wants it: created, so 0 is never mistaken for one.
func program(handle gles.Program) gl.Program { return gl.Program{Init: true, Value: uint32(handle)} }

func (mobile mobileGL) AttachShader(handle gles.Program, shader gles.Shader) {
	mobile.context.AttachShader(program(handle), gl.Shader{Value: uint32(shader)})
}

func (mobile mobileGL) LinkProgram(handle gles.Program) { mobile.context.LinkProgram(program(handle)) }

func (mobile mobileGL) ProgramLinked(handle gles.Program) (bool, string) {
	if mobile.context.GetProgrami(program(handle), gl.LINK_STATUS) != 0 {
		return true, ""
	}
	return false, mobile.context.GetProgramInfoLog(program(handle))
}

func (mobile mobileGL) BindAttribLocation(handle gles.Program, index uint32, name string) {
	mobile.context.BindAttribLocation(program(handle), gl.Attrib{Value: uint(index)}, name)
}

func (mobile mobileGL) UseProgram(handle gles.Program) { mobile.context.UseProgram(program(handle)) }

func (mobile mobileGL) UniformLocation(handle gles.Program, name string) gles.Uniform {
	return gles.Uniform(mobile.context.GetUniformLocation(program(handle), name).Value)
}

func uniform(location gles.Uniform) gl.Uniform { return gl.Uniform{Value: int32(location)} }

func (mobile mobileGL) Uniform1i(location gles.Uniform, value int32) {
	mobile.context.Uniform1i(uniform(location), int(value))
}

func (mobile mobileGL) Uniform1f(location gles.Uniform, value float32) {
	mobile.context.Uniform1f(uniform(location), value)
}

func (mobile mobileGL) Uniform2f(location gles.Uniform, first, second float32) {
	mobile.context.Uniform2f(uniform(location), first, second)
}

func (mobile mobileGL) Uniform3f(location gles.Uniform, first, second, third float32) {
	mobile.context.Uniform3f(uniform(location), first, second, third)
}

func (mobile mobileGL) Uniform4f(location gles.Uniform, first, second, third, fourth float32) {
	mobile.context.Uniform4f(uniform(location), first, second, third, fourth)
}

func (mobile mobileGL) UniformMatrix3(location gles.Uniform, matrix [9]float32) {
	mobile.context.UniformMatrix3fv(uniform(location), matrix[:])
}

func (mobile mobileGL) UniformMatrix4(location gles.Uniform, matrix [16]float32) {
	mobile.context.UniformMatrix4fv(uniform(location), matrix[:])
}

func (mobile mobileGL) CreateBuffer() gles.Buffer {
	return gles.Buffer(mobile.context.CreateBuffer().Value)
}

func (mobile mobileGL) BindBuffer(target gles.Enum, buffer gles.Buffer) {
	mobile.context.BindBuffer(gl.Enum(target), gl.Buffer{Value: uint32(buffer)})
}

func (mobile mobileGL) BufferData(target gles.Enum, data []byte, usage gles.Enum) {
	mobile.context.BufferData(gl.Enum(target), data, gl.Enum(usage))
}

func (mobile mobileGL) EnableVertexAttribArray(index uint32) {
	mobile.context.EnableVertexAttribArray(gl.Attrib{Value: uint(index)})
}

func (mobile mobileGL) VertexAttribPointer(index uint32, size int, kind gles.Enum, normalized bool, stride, offset int) {
	mobile.context.VertexAttribPointer(gl.Attrib{Value: uint(index)}, size, gl.Enum(kind), normalized, stride, offset)
}

func (mobile mobileGL) CreateTexture() gles.Texture {
	return gles.Texture(mobile.context.CreateTexture().Value)
}

func (mobile mobileGL) ActiveTexture(unit gles.Enum) { mobile.context.ActiveTexture(gl.Enum(unit)) }

func (mobile mobileGL) BindTexture(target gles.Enum, texture gles.Texture) {
	mobile.context.BindTexture(gl.Enum(target), gl.Texture{Value: uint32(texture)})
}

func (mobile mobileGL) TexParameteri(target, name gles.Enum, value int32) {
	mobile.context.TexParameteri(gl.Enum(target), gl.Enum(name), int(value))
}

func (mobile mobileGL) PixelStorei(name gles.Enum, value int32) {
	mobile.context.PixelStorei(gl.Enum(name), value)
}

func (mobile mobileGL) TexImage2D(target gles.Enum, level int, internalFormat gles.Enum, width, height int, format, kind gles.Enum, pixels []byte) {
	mobile.context.TexImage2D(gl.Enum(target), level, int(internalFormat), width, height, gl.Enum(format), gl.Enum(kind), pixels)
}

func (mobile mobileGL) Viewport(x, y, width, height int) {
	mobile.context.Viewport(x, y, width, height)
}

func (mobile mobileGL) ClearColor(red, green, blue, alpha float32) {
	mobile.context.ClearColor(red, green, blue, alpha)
}

func (mobile mobileGL) Clear(mask gles.Enum) { mobile.context.Clear(gl.Enum(mask)) }

func (mobile mobileGL) Enable(capability gles.Enum) { mobile.context.Enable(gl.Enum(capability)) }

func (mobile mobileGL) Disable(capability gles.Enum) { mobile.context.Disable(gl.Enum(capability)) }

func (mobile mobileGL) BlendFuncSeparate(sourceColor, destinationColor, sourceAlpha, destinationAlpha gles.Enum) {
	mobile.context.BlendFuncSeparate(gl.Enum(sourceColor), gl.Enum(destinationColor), gl.Enum(sourceAlpha), gl.Enum(destinationAlpha))
}

func (mobile mobileGL) DepthMask(write bool) { mobile.context.DepthMask(write) }

func (mobile mobileGL) DepthFunc(function gles.Enum) { mobile.context.DepthFunc(gl.Enum(function)) }

func (mobile mobileGL) DrawArrays(mode gles.Enum, first, count int) {
	mobile.context.DrawArrays(gl.Enum(mode), first, count)
}

func (mobile mobileGL) DrawElements(mode gles.Enum, count int, kind gles.Enum, offset int) {
	mobile.context.DrawElements(gl.Enum(mode), count, gl.Enum(kind), offset)
}
