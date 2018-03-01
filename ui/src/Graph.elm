module Graph
    exposing
        ( Config
        , Model
        , Msg(Draw)
        , new
        , subscriptions
        , update
        , view
        )

import Api
import Html exposing (Html)
import Html.Attributes as Attr
import Http
import Process
import Task
import Time exposing (Time)
import URLQuery exposing (URLQuery)
import Window


-- CONFIG


type alias Config model msg =
    { get : model -> Model
    , set : model -> Model -> model
    , wrap : Msg -> msg
    }



-- MODEL


type alias Info =
    { metric : String
    , width : Int
    }


type State
    = Nothing
    | Delay String
    | Image Info


type Model
    = Model
        { state : State
        , size : Maybe Window.Size -- size of the window
        , counter : Int
        }


new : Config model msg -> ( Model, Cmd msg )
new config =
    ( Model
        { state = Nothing
        , size = Maybe.Nothing
        , counter = 0
        }
    , Cmd.map config.wrap <|
        Task.perform Resize Window.size
    )



-- MSG


type Msg
    = Draw String
    | Resize Window.Size
    | ResizeTimer Int



-- SUBSCRIPTIONS


subscriptions : Config model msg -> model -> Sub msg
subscriptions config model =
    doSubscriptions (config.get model)
        |> Sub.map config.wrap


doSubscriptions : Model -> Sub Msg
doSubscriptions (Model model) =
    Window.resizes Resize



-- UPDATE


update : Config model msg -> model -> Msg -> ( model, Cmd msg )
update config model msg =
    doUpdate config (config.get model) msg
        |> Tuple.mapFirst (config.set model)
        |> Tuple.mapSecond (Cmd.map config.wrap)


delay : Time
delay =
    250 * Time.millisecond


doUpdate : Config model msg -> Model -> Msg -> ( Model, Cmd Msg )
doUpdate config (Model model) msg =
    case msg of
        Draw metric ->
            ( case model.size of
                Maybe.Nothing ->
                    Model { model | state = Delay metric }

                Just { width } ->
                    Model { model | state = Image (Info metric width) }
            , Cmd.none
            )

        Resize size ->
            let
                newCounter =
                    model.counter + 1
            in
            ( Model { model | size = Just size, counter = newCounter }
            , sendMessageAfter delay (ResizeTimer newCounter)
            )

        ResizeTimer counter ->
            ( case ( counter == model.counter, model.state, model.size ) of
                ( True, Delay metric, Just { width } ) ->
                    Model { model | state = Image (Info metric width) }

                ( True, Image info, Just { width } ) ->
                    Model { model | state = Image (Info info.metric width) }

                _ ->
                    Model model
            , Cmd.none
            )



-- VIEW


view : Config model msg -> Model -> Html msg
view config model =
    doView model
        |> Html.map config.wrap


doView : Model -> Html Msg
doView (Model model) =
    case model.state of
        Image { metric, width } ->
            let
                query =
                    URLQuery.empty
                        |> URLQuery.add "metric" metric
                        |> URLQuery.add "width" (toString (width - 10))
                        |> URLQuery.add "padding" "5"
                        |> URLQuery.render
            in
            Html.img [ Attr.src <| "/api/render" ++ query ] []

        _ ->
            Html.text ""



-- UTILS


sendMessage : msg -> Cmd msg
sendMessage msg =
    Task.perform (always msg) (Task.succeed ())


sendMessageAfter : Time -> msg -> Cmd msg
sendMessageAfter delay msg =
    Task.perform (always msg) (Process.sleep delay)
