package main

import (
	"fmt"
	"github.com/scottferg/Fergulator/nes"
	"github.com/scottferg/Go-SDL/gfx"
	"github.com/scottferg/Go-SDL/sdl"
	"github.com/go-gl/gl/v2.1/gl"
	"log"
	"math"
	"os"
	"unsafe"
)

type Video struct {
	videoTick     <-chan []uint32
	screen        *sdl.Surface
	fpsmanager    *gfx.FPSmanager
	prog          uint32
	texture       uint32
	width, height int32
	textureUni    int32
	Fullscreen    bool
}

func createProgram(vertShaderSrc string, fragShaderSrc string) uint32 {
	vertShader := loadShader(gl.VERTEX_SHADER, vertShaderSrc)
	fragShader := loadShader(gl.FRAGMENT_SHADER, fragShaderSrc)

	prog := gl.CreateProgram()

	gl.AttachShader(prog,vertShader)
	gl.AttachShader(prog,fragShader)
	gl.LinkProgram(prog)

	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status != gl.TRUE {
		//log := gl.GetInfoLogARB(prog)
		panic(fmt.Errorf("Failed to link program: , "))
	}

	return prog
}

func loadShader(shaderType uint32, source string) uint32 {
	shader := gl.CreateShader(shaderType)
	if err := gl.GetError(); err != gl.NO_ERROR {
		panic(fmt.Errorf("gl error: %v", err))
	}

	csource := gl.Str(source)
	gl.ShaderSource(shader, 1, &csource, nil)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status != gl.TRUE {
		//log := shader.GetInfoLog()
		panic(fmt.Errorf("Failed to compile shader:  shader: %v", source))
	}

	return shader
}

func (v *Video) Init(t <-chan []uint32, n string) {
	v.videoTick = t

	if sdl.Init(sdl.INIT_VIDEO|sdl.INIT_JOYSTICK|sdl.INIT_AUDIO) != 0 {
		log.Fatal(sdl.GetError())
	}

	v.screen = sdl.SetVideoMode(512, 480, 32,
		sdl.OPENGL|sdl.RESIZABLE|sdl.GL_DOUBLEBUFFER)
	if v.screen == nil {
		log.Fatal(sdl.GetError())
	}

	sdl.WM_SetCaption(fmt.Sprintf("Fergulator - %s", n), "")

	v.initGL()
	v.Reshape(v.screen.W, v.screen.H)

	v.fpsmanager = gfx.NewFramerate()
	v.fpsmanager.SetFramerate(60)

	return
}

func (v *Video) initGL() {
	if err := gl.Init(); err != nil {
		panic(err)
	}

	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	v.prog = createProgram(vertShaderSrcDef, fragShaderSrcDef)
	posAttrib := uint32(gl.GetAttribLocation(v.prog,gl.Str("vPosition"+ "\x00")))
	texCoordAttr := uint32(gl.GetAttribLocation(v.prog,gl.Str("vTexCoord"+ "\x00")))
	v.textureUni = gl.GetAttribLocation(v.prog,gl.Str("texture"+ "\x00"))

  	var texture uint32
  	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D,texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.UseProgram(v.prog)
	gl.EnableVertexAttribArray(posAttrib)
	gl.EnableVertexAttribArray(texCoordAttr)
	//posAttrib.EnableArray()
	//texCoordAttr.EnableArray()

	var vbo uint32
   	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	verts := []float32{-1.0, 1.0, -1.0, -1.0, 1.0, -1.0, 1.0, -1.0, 1.0, 1.0, -1.0, 1.0}
	gl.BufferData(gl.ARRAY_BUFFER, len(verts)*int(unsafe.Sizeof(verts[0])), gl.Ptr(verts), gl.STATIC_DRAW)

	var textCoorBuf uint32
	gl.GenBuffers(1,&textCoorBuf)
	gl.BindBuffer(gl.ARRAY_BUFFER, textCoorBuf)
	texVerts := []float32{0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 1.0, 0.0, 1.0}
	gl.BufferData(gl.ARRAY_BUFFER, len(texVerts)*int(unsafe.Sizeof(texVerts[0])), gl.Ptr(texVerts), gl.STATIC_DRAW)


	gl.VertexAttribPointer(posAttrib,2, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.VertexAttribPointer(texCoordAttr,2, gl.FLOAT, false, 0, gl.PtrOffset(0))
	//posAttrib.AttribPointer(2, gl.FLOAT, false, 0, uintptr(0))
	//texCoordAttr.AttribPointer(2, gl.FLOAT, false, 0, uintptr(0))
}

func (v *Video) ResizeEvent(w, h int32) {
	v.screen = sdl.SetVideoMode(int(w), int(h), 32, sdl.OPENGL|sdl.RESIZABLE)
	v.Reshape(w, h)
}

func (v *Video) FullscreenEvent() {
	v.screen = sdl.SetVideoMode(1440, 900, 32, sdl.OPENGL|sdl.FULLSCREEN)
	v.Reshape(1440, 900)
}

func (v *Video) Reshape(width int32, height int32) {
	var x_offset int32 = 0
	var y_offset int32 = 0

	r := ((float64)(height)) / ((float64)(width))

	if r > 0.9375 { // Height taller than ratio
		h := (int32)(math.Floor((float64)(0.9375 * (float64)(width))))
		y_offset = (height - h) / 2
		height = h
	} else if r < 0.9375 { // Width wider
		w := (int32)(math.Floor((float64)((256.0 / 240.0) * (float64)(height))))
		x_offset = (width - w) / 2
		width = w
	}

	v.width = width
	v.height = height

	gl.Viewport(x_offset, y_offset, width, height)
}

func quit_event() int {
	running = false
	return 0
}

func (v *Video) Render() {
	for running {
		select {
		case buf := <-v.videoTick:
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			gl.UseProgram(v.prog)

			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(v.texture,gl.TEXTURE_2D)

			gl.TexImage2D(gl.TEXTURE_2D, 0, 3, 240, 224, 0, gl.RGBA,
				gl.UNSIGNED_INT_8_8_8_8, gl.Ptr(buf))

			gl.DrawArrays(gl.TRIANGLES, 0, 6)

			if v.screen != nil {
				sdl.GL_SwapBuffers()
				v.fpsmanager.FramerateDelay()
			}
		case ev := <-sdl.Events:
			switch e := ev.(type) {
			case sdl.ResizeEvent:
				v.ResizeEvent(e.W, e.H)
			case sdl.QuitEvent:
				os.Exit(0)
			case sdl.KeyboardEvent:
				switch e.Keysym.Sym {
				case sdl.K_ESCAPE:
					running = false
				case sdl.K_r:
					// Trigger reset interrupt
					if e.Type == sdl.KEYDOWN {
						// cpu.RequestInterrupt(InterruptReset)
					}
				case sdl.K_l:
					if e.Type == sdl.KEYDOWN {
						nes.LoadGameState()
					}
				case sdl.K_s:
					if e.Type == sdl.KEYDOWN {
						nes.SaveGameState()
					}
				case sdl.K_i:
					if e.Type == sdl.KEYDOWN {
						nes.AudioEnabled = !nes.AudioEnabled
					}
				case sdl.K_p:
					if e.Type == sdl.KEYDOWN {
						nes.TogglePause()
					}
				case sdl.K_d:
					if e.Type == sdl.KEYDOWN {
						jsHandler.ReloadFile(debugfile)
					}
				case sdl.K_m:
					if e.Type == sdl.KEYDOWN {
						nes.Handler.Handle("debug-mode")
					}
				case sdl.K_BACKSLASH:
					if e.Type == sdl.KEYDOWN {
						nes.Pause()
						nes.StepFrame()
					}
				case sdl.K_1:
					if e.Type == sdl.KEYDOWN {
						v.ResizeEvent(256, 240)
					}
				case sdl.K_2:
					if e.Type == sdl.KEYDOWN {
						v.ResizeEvent(512, 480)
					}
				case sdl.K_3:
					if e.Type == sdl.KEYDOWN {
						v.ResizeEvent(768, 720)
					}
				case sdl.K_4:
					if e.Type == sdl.KEYDOWN {
						v.ResizeEvent(1024, 960)
					}
				}

				switch e.Type {
				case sdl.KEYDOWN:
					nes.Pads[0].KeyDown(e, 0)
				case sdl.KEYUP:
					nes.Pads[0].KeyUp(e, 0)
				}
			}
		}
	}
}

func (v *Video) Close() {
	sdl.Quit()
}
